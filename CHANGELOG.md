# Changelog

All notable changes to Claude Style. Format: [Keep a Changelog](https://keepachangelog.com/),
versions: [SemVer](https://semver.org/). Claude Style forked from claude-squad v1.0.20.

## Unreleased

### Added
- Color themes: `claude-style` (default), `monokai`, `vscode`, `ocean-green`, `dracula`, `nord`,
  `gruvbox`, `solarized`. Cycle with `t` in the app or use `claude-style themes [name]`; saved in
  config as `theme`. Also styles each session's tmux status bar.

### Changed
- Rebranded from claude-squad: binary `claude-style`, config `~/.claude-style`, tmux prefix
  `claudestyle_`. First run migrates `~/.claude-squad` config/state and renames running sessions.
- Branch picker lists branches checked out elsewhere greyed out and unselectable.

### Performance
- Pane previews are streamed from the session's tmux client instead of polling `capture-pane`
  (two idle sessions: 124 tmux processes per 10 s → 0).

### Fixed
- Detaching (Ctrl+Q) from an idle session hung forever, leaving you stuck in full screen:
  closing the PTY doesn't interrupt a blocked read on macOS, so the tmux client is now killed.
- Ctrl+Q was ignored when it arrived in the same read as other input (typed quickly after
  other keys, or with a focus event); also accepts the CSI-u / modifyOtherKeys encodings.
- After detaching, the terminal's cursor stayed visible over the TUI (green block in the
  corner) and focus reporting stayed on.
- First keystroke / start of a paste lost when attaching (TUI no longer reads stdin while attached).
- Background `git fetch` could prompt for credentials on the TTY and hang; now never prompts, 20 s timeout.
- Prompt from Shift+N could be dropped while the agent was still booting.
- New session silently vanished when its branch was already checked out elsewhere; now a clear error.
- Terminal tab hid the top row of the pane (a fresh shell looked empty).
- Preview replaced the pane's last row with `...`.
- Tests no longer touch the real `~/.claude-style` or the default tmux server.
