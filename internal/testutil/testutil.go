// Package testutil keeps tests away from the developer's real state: ~/.claude-style,
// git worktrees under it, and the default tmux server with live sessions.
package testutil

import (
	"os"
	"path/filepath"
)

// Isolate points HOME and TMUX_TMPDIR at a fresh temp dir. Call it first in TestMain and
// run the returned cleanup after m.Run.
func Isolate() (cleanup func()) {
	dir, err := os.MkdirTemp("", "cstest")
	if err != nil {
		panic(err)
	}
	home := filepath.Join(dir, "home")
	tmuxDir := filepath.Join(dir, "tmux") // kept short: tmux socket paths are limited to ~104 bytes
	for _, d := range []string{home, tmuxDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			panic(err)
		}
	}
	os.Setenv("HOME", home)
	os.Setenv("TMUX_TMPDIR", tmuxDir)
	os.Unsetenv("TMUX")
	return func() { _ = os.RemoveAll(dir) }
}
