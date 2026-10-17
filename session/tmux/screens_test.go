package tmux

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"claude-booths/cmd/cmd_test"

	"github.com/stretchr/testify/require"
)

// Screens as the real agents draw them (Claude Code 2.1.299, Gemini CLI 0.64, Codex 0.146, Aider 0.86).
const (
	claudeTrustNo = ` Accessing workspace:
 /private/var/folders/xx/T/tmp.LVPKU9m77f
 Quick safety check: Is this a project you created or one you trust?
 Claude Code'll be able to read, edit, and execute files here.
 Security guide
 ❯ No, exit
   Yes, I trust this folder
 Enter to confirm · Esc to cancel
`
	claudeTrustYes = ` Quick safety check: Is this a project you created or one you trust?
   No, exit
 ❯ Yes, I trust this folder
 Enter to confirm · Esc to cancel
`
	claudeInput = ` ▐▛███▛█   Claude Code v2.1.299
────────────────────────────────
❯ Try "fix lint errors"
────────────────────────────────
  ⏵⏵ don't ask on (shift+tab to cycle)
`
	claudePermission = `❯ Change x to 2 in a.ts
 Edit file
 a.ts
 1 -export const x = 1;
 1 +export const x = 2;
 Do you want to make this edit to a.ts?
 ❯ 1. Yes
   2. Yes, and switch to accept edits (shift+tab)
   3. No
 Esc to cancel · Tab to amend
`
	geminiTrust = ` │ Do you trust the files in this folder?                     │
 │ Trusting a folder allows Gemini CLI to load its local configurations │
 │ ● 1. Trust folder (tmp.WSiNi8KYH3)                          │
 │   2. Trust parent folder (T)                                │
 │   3. Don't trust                                            │
 ╰─────────────────────────────────────────────────────────────╯
`
	codexUpdate = `  ✨ Update available! 0.146.0 -> 0.162.1
  Release notes: https://github.com/openai/codex/releases/latest
› 1. Update now (runs ` + "`sh -c 'curl -fsSL https://chatgpt.com/codex/install.sh | sh'`" + `)
  2. Skip
  3. Skip until next version
  Press enter to continue
`
	codexTrust = `> You are in /private/var/folders/xx/T/tmp.WSiNi8KYH3
  Do you trust the contents of this directory? Working with untrusted contents comes with higher risk.
› 1. Yes, continue
  2. No, quit
  Press enter to continue
`
	aiderGitignore = `You can skip this check with --no-gitignore
Add .aider* to .gitignore (recommended)? (Y)es/(N)o [Yes]: `
	aiderDocs = `Warning: ollama_chat/qwen2.5-coder:7b expects these environment variables
Open documentation url for more info? (Y)es/(N)o/(D)on't ask again [Yes]: `
	aiderAnswered = `Add .aider* to .gitignore (recommended)? (Y)es/(N)o [Yes]: n
Aider v0.86.2
Model: ollama_chat/qwen2.5-coder:7b with whole edit format
───────────────────────────────
> `
)

func TestMatchScreen(t *testing.T) {
	cases := []struct {
		name, content         string
		agent, choose, keys   string
		blocked, wantNoScreen bool
	}{
		{name: "claude trust", content: claudeTrustNo, agent: "Claude Code", choose: "Yes, I trust this folder"},
		{name: "gemini trust", content: geminiTrust, agent: "Gemini CLI", choose: "Trust folder ("},
		{name: "codex trust", content: codexTrust, agent: "Codex", choose: "Yes, continue"},
		{name: "codex update is never accepted", content: codexUpdate, agent: "Codex", blocked: true},
		{name: "aider gitignore", content: aiderGitignore, agent: "Aider", keys: "n\r"},
		{name: "aider docs", content: aiderDocs, agent: "Aider", keys: "d\r"},
		{name: "unknown y/n question", content: "Something new? (Y)es/(N)o [Yes]: ", agent: "The agent", blocked: true},
		{name: "amp lowercase y/n question", content: "No API key found. Would you like to log in? [(y)es, (n)o]: ", agent: "The agent", blocked: true},
		{name: "gemini sign-in", content: " │ How would you like to authenticate for this project? │\n │ ● 1. Sign in with Google │\n", agent: "Gemini CLI", blocked: true},
		{name: "claude input box", content: claudeInput, wantNoScreen: true},
		{name: "claude permission prompt is not a startup screen", content: claudePermission, wantNoScreen: true},
		{name: "answered aider question in scrollback", content: aiderAnswered, wantNoScreen: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := matchScreen(c.content)
			if c.wantNoScreen {
				require.Nil(t, s)
				return
			}
			require.NotNil(t, s)
			require.Equal(t, c.agent, s.agent)
			require.Equal(t, c.blocked, s.block != "")
			require.Equal(t, c.choose, s.choose)
			require.Equal(t, c.keys, s.keys)
		})
	}
}

func TestSelectedOption(t *testing.T) {
	// capture-pane -e keeps colors; the cursor is usually colored.
	require.Contains(t, selectedOption("   No, exit\n \x1b[38;5;153m❯\x1b[39m Yes, I trust this folder\n"), "Yes, I trust")
	require.Contains(t, selectedOption(claudeTrustNo), "No, exit")
	require.Contains(t, selectedOption(claudeTrustYes), "Yes, I trust this folder")
	require.Contains(t, selectedOption(geminiTrust), "Trust folder (")
	require.Contains(t, selectedOption(codexTrust), "Yes, continue")
}

func TestHasPermissionPrompt(t *testing.T) {
	// Wrappers, paths and flags used to switch detection off (it required program == "claude").
	for _, prog := range []string{"claude", "/Users/me/.local/bin/claude-trusted", "claude --model opus"} {
		require.True(t, hasPermissionPrompt(prog, claudePermission), prog)
		require.False(t, hasPermissionPrompt(prog, claudeInput), prog)
	}
	require.True(t, hasPermissionPrompt("aider --model x", aiderGitignore))
	// The question is still in scrollback but aider is back at its input: don't press Enter again.
	require.False(t, hasPermissionPrompt("aider --model x", aiderAnswered))
	require.False(t, hasPermissionPrompt("codex", claudePermission))
}

// screenSession is a started TmuxSession whose pane shows screen(keys written so far).
func screenSession(t *testing.T, program string, screen func(written string) string) (*TmuxSession, func() string) {
	t.Helper()
	oldEvery, oldPolls := stableEvery, stablePolls
	stableEvery, stablePolls = 5*time.Millisecond, 2
	t.Cleanup(func() { stableEvery, stablePolls = oldEvery, oldPolls })

	ptyFactory := NewMockPtyFactory(t)
	var mu sync.Mutex
	created := false
	written := func() string {
		mu.Lock()
		defer mu.Unlock()
		if len(ptyFactory.files) < 2 {
			return ""
		}
		b, _ := os.ReadFile(ptyFactory.files[len(ptyFactory.files)-1].Name())
		return string(b)
	}
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			if strings.Contains(cmd.String(), "has-session") && !created {
				created = true
				return fmt.Errorf("no session yet")
			}
			return nil
		},
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) {
			if strings.Contains(cmd.String(), "capture-pane") {
				return []byte(screen(written())), nil
			}
			return nil, nil
		},
	}
	s := newTmuxSession("screens", program, ptyFactory, cmdExec)
	require.NoError(t, s.Start(t.TempDir()))
	return s, written
}

func TestHandleStartupScreensAcceptsClaudeTrust(t *testing.T) {
	// Before: nothing handled this screen, so the prompt's Enter picked "No, exit" and the booth died.
	s, written := screenSession(t, "claude", func(w string) string {
		switch {
		case strings.Contains(w, "\x1b[B\r"):
			return claudeInput
		case strings.Contains(w, "\x1b[B"):
			return claudeTrustYes
		}
		return claudeTrustNo
	})
	require.NoError(t, s.HandleStartupScreens(2*time.Second))
	require.Equal(t, "\x1b[B\r", written(), "move to Yes, then Enter")
}

func TestHandleStartupScreensRefusesCodexUpdate(t *testing.T) {
	s, written := screenSession(t, "codex", func(string) string { return codexUpdate })
	err := s.HandleStartupScreens(2 * time.Second)
	var se *SetupScreenError
	require.True(t, errors.As(err, &se), "got %v", err)
	require.Equal(t, "Codex", se.Agent)
	require.Empty(t, written(), "nothing may be typed into the update screen")
}

func TestHandleStartupScreensAnswersAiderQuestions(t *testing.T) {
	s, written := screenSession(t, "aider", func(w string) string {
		switch strings.Count(w, "\r") {
		case 0:
			return aiderGitignore
		case 1:
			return aiderDocs
		}
		return aiderAnswered
	})
	require.NoError(t, s.HandleStartupScreens(2*time.Second))
	require.Equal(t, "n\rd\r", written())
}

func TestAutoYesAnswersEachPromptOnce(t *testing.T) {
	// Aider keeps the answered question on screen until it redraws; auto-yes used to keep
	// pressing Enter on it.
	s, _ := screenSession(t, "aider", func(string) string { return aiderGitignore })
	_, hasPrompt := s.HasUpdated()
	require.True(t, hasPrompt)
	require.NoError(t, s.TapEnter())
	_, hasPrompt = s.HasUpdated()
	require.False(t, hasPrompt, "same screen after Enter: already answered")
}

func TestHandleStartupScreensWaitsForTheAgentToDraw(t *testing.T) {
	// Before: a blank pane counted as settled, so the prompt was typed before the agent started
	// and lost (Gemini, Aider, slow OpenCode).
	captures := 0
	s, written := screenSession(t, "gemini", func(string) string {
		captures++
		if captures < 10 {
			return "\n\n"
		}
		return claudeInput
	})
	require.NoError(t, s.HandleStartupScreens(2*time.Second))
	require.GreaterOrEqual(t, captures, 10, "returned while the pane was still blank")
	require.Empty(t, written())
}

func TestHandleStartupScreensBlankPaneIsAnError(t *testing.T) {
	s, written := screenSession(t, "gemini", func(string) string { return "" })
	require.Error(t, s.HandleStartupScreens(100*time.Millisecond))
	require.Empty(t, written())
}
