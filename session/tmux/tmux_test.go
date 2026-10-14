package tmux

import (
	cmd2 "claude-booths/cmd"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"claude-booths/cmd/cmd_test"

	"github.com/stretchr/testify/require"
)

type MockPtyFactory struct {
	t *testing.T

	// Array of commands and the corresponding file handles representing PTYs.
	cmds  []*exec.Cmd
	files []*os.File
}

func (pt *MockPtyFactory) Start(cmd *exec.Cmd) (*os.File, error) {
	filePath := filepath.Join(pt.t.TempDir(), fmt.Sprintf("pty-%s-%d", pt.t.Name(), rand.Int31()))
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err == nil {
		pt.cmds = append(pt.cmds, cmd)
		pt.files = append(pt.files, f)
	}
	return f, err
}

func (pt *MockPtyFactory) Close() {}

func NewMockPtyFactory(t *testing.T) *MockPtyFactory {
	return &MockPtyFactory{
		t: t,
	}
}

func TestSanitizeName(t *testing.T) {
	session := NewTmuxSession("asdf", "program")
	require.Equal(t, TmuxPrefix+"asdf", session.sanitizedName)

	session = NewTmuxSession("a sd f . . asdf", "program")
	require.Equal(t, TmuxPrefix+"asdf__asdf", session.sanitizedName)
}

func TestStartTmuxSession(t *testing.T) {
	ptyFactory := NewMockPtyFactory(t)

	created := false
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			if strings.Contains(cmd.String(), "has-session") && !created {
				created = true
				return fmt.Errorf("session already exists")
			}
			return nil
		},
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) {
			return []byte("output"), nil
		},
	}

	workdir := t.TempDir()
	session := newTmuxSession("test-session", "claude", ptyFactory, cmdExec)

	err := session.Start(workdir)
	require.NoError(t, err)
	require.Equal(t, 2, len(ptyFactory.cmds))
	require.Equal(t, fmt.Sprintf("tmux new-session -d -s claudebooths_test-session -c %s claude", workdir),
		cmd2.ToString(ptyFactory.cmds[0]))
	require.Equal(t, "tmux attach-session -t claudebooths_test-session",
		cmd2.ToString(ptyFactory.cmds[1]))

	require.Equal(t, 2, len(ptyFactory.files))

	// File should be closed.
	_, err = ptyFactory.files[0].Stat()
	require.Error(t, err)
	// File should be open
	_, err = ptyFactory.files[1].Stat()
	require.NoError(t, err)
}

// A tmux server that has gone away (reboot, crash, `tmux kill-server`) takes every session
// with it. attach-session against a missing session still forks successfully, so Restore
// has to check for the session itself or it reports success while attached to nothing.
func TestRestoreReturnsErrSessionNotFoundWhenSessionIsGone(t *testing.T) {
	ptyFactory := NewMockPtyFactory(t)
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			if strings.Contains(cmd.String(), "has-session") {
				return fmt.Errorf("can't find session")
			}
			return nil
		},
	}

	session := NewTmuxSessionWithDeps("gone", "program", ptyFactory, cmdExec)
	err := session.Restore()

	require.ErrorIs(t, err, ErrSessionNotFound)
	require.Empty(t, ptyFactory.cmds, "should not have opened a PTY for a session that does not exist")
}

func TestRestoreAttachesWhenSessionExists(t *testing.T) {
	ptyFactory := NewMockPtyFactory(t)
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error { return nil },
	}

	session := NewTmuxSessionWithDeps("alive", "program", ptyFactory, cmdExec)
	require.NoError(t, session.Restore())
	require.Len(t, ptyFactory.cmds, 1)
	require.Contains(t, ptyFactory.cmds[0].String(), "attach-session")
}

// While the pump sees no output, CapturePaneContent serves the cached capture instead of
// spawning tmux; any output (or the pump dying) makes it capture again.
func TestCapturePaneContentStreamsInsteadOfPolling(t *testing.T) {
	captures := 0
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error { return nil },
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) {
			captures++
			return []byte(fmt.Sprintf("screen %d", captures)), nil
		},
	}
	session := NewTmuxSessionWithDeps("stream", "program", NewMockPtyFactory(t), cmdExec)
	r, w, err := os.Pipe()
	require.NoError(t, err)
	session.startPump(r)

	waitFor := func(cond func() bool) {
		deadline := time.Now().Add(2 * time.Second)
		for !cond() && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
		}
	}

	got, _ := session.CapturePaneContent()
	require.Equal(t, "screen 1", got)
	for i := 0; i < 5; i++ {
		got, _ = session.CapturePaneContent()
	}
	require.Equal(t, "screen 1", got, "idle pane must come from cache")
	require.Equal(t, 1, captures)

	_, _ = w.Write([]byte("x"))
	waitFor(session.dirty.Load)
	got, _ = session.CapturePaneContent()
	require.Equal(t, "screen 2", got, "output must trigger a fresh capture")

	_ = w.Close() // pump exits: no streaming, so every call captures again
	waitFor(func() bool { return !session.streaming.Load() })
	_, _ = session.CapturePaneContent()
	_, _ = session.CapturePaneContent()
	require.Equal(t, 4, captures)
}

func TestDetachKeyIndex(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{"\x11", 0},
		{"abc\x11", 3},        // typed right after other keys, arrives in one read
		{"\x1b[I\x11", 3},     // after a focus-in event
		{"x\x1b[113;5u", 1},   // CSI-u (kitty keyboard / extended keys)
		{"\x1b[27;5;113~", 0}, // xterm modifyOtherKeys
		{"hello", -1},
		{"\x1b[113u", -1}, // plain q in CSI-u, no ctrl
	} {
		if got := detachKeyIndex([]byte(tc.in)); got != tc.want {
			t.Errorf("detachKeyIndex(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// Regression: closing a pty master doesn't unblock a pending Read on macOS, so detaching from
// an idle session (no output) hung forever waiting for the pump. stopClient must kill the
// client so the pump ends promptly.
func TestStopClientEndsPumpOfIdleSession(t *testing.T) {
	client := exec.Command("sleep", "30") // stands in for an idle `tmux attach-session`
	ptmx, err := MakePtyFactory().Start(client)
	require.NoError(t, err)
	session := NewTmuxSessionWithDeps("idle", "program", NewMockPtyFactory(t), cmd_test.MockCmdExec{})
	session.ptmx, session.client = ptmx, client
	session.startPump(ptmx)
	time.Sleep(100 * time.Millisecond) // let the pump block in Read

	start := time.Now()
	require.NoError(t, session.stopClient())
	select {
	case <-session.pumpDone:
	default:
		t.Fatal("pump still running after stopClient")
	}
	require.Less(t, time.Since(start), 2*time.Second, "stopClient waited for its timeout instead of ending the pump")
	require.False(t, session.streaming.Load())
}

// Deleting a booth whose agent already exited must not fail on its gone tmux session, and
// must never fall back to tmux's prefix match (which could kill another booth's session).
func TestCloseSessionAlreadyGone(t *testing.T) {
	var ran []string
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			ran = append(ran, cmd.String())
			if strings.Contains(cmd.String(), "has-session") {
				return fmt.Errorf("exit status 1")
			}
			return nil
		},
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) { return nil, nil },
	}
	session := NewTmuxSessionWithDeps("gone", "program", NewMockPtyFactory(t), cmdExec)
	require.NoError(t, session.Close())
	for _, c := range ran {
		require.NotContains(t, c, "kill-session", "nothing to kill")
	}
}

func TestCloseKillsByExactName(t *testing.T) {
	var kill string
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			if strings.Contains(cmd.String(), "kill-session") {
				kill = cmd.String()
			}
			return nil
		},
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) { return nil, nil },
	}
	session := NewTmuxSessionWithDeps("ab", "program", NewMockPtyFactory(t), cmdExec)
	require.NoError(t, session.Close())
	require.Contains(t, kill, "-t="+TmuxPrefix+"ab")
}

func TestCleanupOrphanTerminalsOnlyKillsTermSessionsWithoutBooth(t *testing.T) {
	var killed []string
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			if strings.Contains(cmd.String(), "kill-session") {
				killed = append(killed, cmd.Args[len(cmd.Args)-1])
			}
			return nil
		},
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) {
			return []byte(strings.Join([]string{
				TmuxPrefix + "alive", TmuxPrefix + "term_alive", // booth exists
				TmuxPrefix + "term_gone", // orphan
				TmuxPrefix + "gone",      // agent session: never touched here
				"work", "term_other",     // not ours
			}, "\n")), nil
		},
	}
	CleanupOrphanTerminals(cmdExec, []string{"alive"})
	require.Equal(t, []string{"-t=" + TmuxPrefix + "term_gone"}, killed)
}

func TestGone(t *testing.T) {
	exists := true
	calls := 0
	cmdExec := cmd_test.MockCmdExec{
		RunFunc: func(cmd *exec.Cmd) error {
			if strings.Contains(cmd.String(), "has-session") {
				calls++
				if !exists {
					return fmt.Errorf("exit status 1")
				}
			}
			return nil
		},
		OutputFunc: func(cmd *exec.Cmd) ([]byte, error) { return nil, nil },
	}
	s := NewTmuxSessionWithDeps("g", "program", NewMockPtyFactory(t), cmdExec)

	s.streaming.Store(true)
	require.False(t, s.Gone())
	require.Zero(t, calls, "no tmux call while the stream is alive")

	s.streaming.Store(false)
	require.False(t, s.Gone())
	exists = false
	require.True(t, s.Gone())
}
