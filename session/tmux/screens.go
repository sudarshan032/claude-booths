package tmux

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// startupScreen is a screen an agent shows instead of its input box: a folder-trust menu, a
// first-run question, an update offer. Typing a prompt into one of these picks whatever is
// selected (Claude Code's trust menu starts on "No, exit"), so HandleStartupScreens answers the
// ones that are safe to answer and refuses to type into the rest.
type startupScreen struct {
	agent  string   // for messages
	match  []string // all of these must be on the bottom of the pane
	choose string   // menu: move the cursor onto the option containing this, then Enter
	keys   string   // question: answer by typing this
	block  string   // don't answer; the user has to (shown in the error)
}

// The user chose to run the agent in this booth's worktree, so trusting it is what they asked
// for. Questions are answered with the choice that changes nothing (no .gitignore edit, no
// analytics, no browser). Anything that installs or updates software is left to the user.
var startupScreens = []startupScreen{
	{agent: "Claude Code", match: []string{"Yes, I trust this folder"}, choose: "Yes, I trust this folder"},
	{agent: "Claude Code", match: []string{"Do you trust the files in this folder?", "Yes, proceed"}, choose: "Yes, proceed"},
	{agent: "Gemini CLI", match: []string{"Do you trust the files in this folder?", "Trust folder ("}, choose: "Trust folder ("},
	{agent: "Codex", match: []string{"Do you trust the contents of this directory?", "Yes, continue"}, choose: "Yes, continue"},
	{agent: "Codex", match: []string{"Update available!", "Update now"}, block: "it offers to update itself"},
	// Sign-in and first-run setup: the user has to do these once, outside any prompt.
	{agent: "Claude Code", match: []string{"Choose the text style"}, block: "first-run setup; run claude once to finish it"},
	{agent: "Claude Code", match: []string{"Select login method"}, block: "it needs you to log in"},
	{agent: "Codex", match: []string{"Sign in with ChatGPT"}, block: "it needs you to sign in (codex login)"},
	{agent: "Gemini CLI", match: []string{"How would you like to authenticate"}, block: "it needs you to sign in"},
	{agent: "Aider", match: []string{"Open documentation url for more info?"}, keys: "d\r"},
	{agent: "Aider", match: []string{"Add .aider* to .gitignore"}, keys: "n\r"},
	{agent: "Aider", match: []string{"anonymous analytics"}, keys: "n\r"},
	{agent: "Aider", match: []string{"what's new in this version"}, keys: "n\r"},
}

// SetupScreenError means the agent is waiting on a screen claude-booths won't answer for you.
type SetupScreenError struct{ Agent, Reason string }

func (e *SetupScreenError) Error() string {
	return fmt.Sprintf("%s is waiting on a setup screen (%s); step inside the booth to answer it", e.Agent, e.Reason)
}

// paneBottom returns the last n non-blank lines of a pane capture, without colors: where agents
// draw their current menu or question. Older text further up (an answered question) is ignored.
func paneBottom(content string, n int) []string {
	var out []string
	lines := strings.Split(ansi.Strip(content), "\n")
	for i := len(lines) - 1; i >= 0 && len(out) < n; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			out = append(out, lines[i])
		}
	}
	for l, r := 0, len(out)-1; l < r; l, r = l+1, r-1 {
		out[l], out[r] = out[r], out[l]
	}
	return out
}

const screenLines = 20 // bottom lines searched for menus and questions

func matchScreen(content string) *startupScreen {
	bottom := strings.Join(paneBottom(content, screenLines), "\n")
	last := strings.Join(paneBottom(content, 1), "")
	for i := range startupScreens {
		s := &startupScreens[i]
		where := bottom
		if s.keys != "" {
			where = last // a question is only pending while it's the last line
		}
		ok := true
		for _, m := range s.match {
			if !strings.Contains(where, m) {
				ok = false
				break
			}
		}
		if ok {
			return s
		}
	}
	// Any other y/n question still waiting on the last line ("(Y)es/(N)o", Amp's
	// "[(y)es, (n)o]"): don't type a prompt into it.
	if strings.Contains(strings.ToLower(last), "(y)es") {
		return &startupScreen{agent: "The agent", block: "it is asking a yes/no question"}
	}
	return nil
}

// selectedOption returns the menu line under the cursor (❯, › or ●), or "".
func selectedOption(content string) string {
	for _, l := range paneBottom(content, screenLines) {
		t := strings.TrimLeft(strings.TrimSpace(l), "│ ")
		for _, c := range []string{"❯", "›", "●"} {
			if strings.HasPrefix(t, c) {
				return t
			}
		}
	}
	return ""
}

// Stable-pane detection: unchanged for stablePolls captures stableEvery apart. Vars for tests.
var (
	stableEvery = 250 * time.Millisecond
	stablePolls = 4
)

// waitStable waits until the agent has drawn something and the pane stops changing (or
// deadline), and returns its content. A blank pane doesn't count as settled: agents can take
// seconds (a minute on a slow VM) before drawing anything, and a prompt typed then is lost.
func (t *TmuxSession) waitStable(deadline time.Time) string {
	prev, _ := t.CapturePaneContent()
	for stable := 0; stable < stablePolls && time.Now().Before(deadline); {
		time.Sleep(stableEvery)
		cur, _ := t.CapturePaneContent()
		switch {
		case cur != prev:
			stable, prev = 0, cur
		case len(paneBottom(cur, 1)) > 0:
			stable++
		}
	}
	return prev
}

// HandleStartupScreens waits for the agent to settle and answers the startup screens it knows
// (see startupScreens) until the agent shows something else, presumably its input box. It
// returns a *SetupScreenError for screens the user has to answer.
func (t *TmuxSession) HandleStartupScreens(maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	for range 8 {
		content := t.waitStable(deadline)
		if len(paneBottom(content, 1)) == 0 {
			return fmt.Errorf("the agent showed nothing within %s", maxWait)
		}
		s := matchScreen(content)
		if s == nil {
			return nil
		}
		if s.block != "" {
			return &SetupScreenError{Agent: s.agent, Reason: s.block}
		}
		if err := t.answerScreen(s, content); err != nil {
			return err
		}
	}
	return nil
}

func (t *TmuxSession) answerScreen(s *startupScreen, content string) error {
	if s.keys != "" {
		return t.SendKeys(s.keys)
	}
	for range 5 {
		if strings.Contains(selectedOption(content), s.choose) {
			return t.TapEnter()
		}
		if err := t.SendKeys("\x1b[B"); err != nil { // down arrow
			return err
		}
		time.Sleep(stableEvery)
		content, _ = t.CapturePaneContent()
	}
	return fmt.Errorf("could not select %q on %s's screen", s.choose, s.agent)
}
