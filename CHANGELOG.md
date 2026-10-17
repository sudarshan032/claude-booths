# Changelog

All notable changes to Claude Booths. Format: [Keep a Changelog](https://keepachangelog.com/),
versions: [SemVer](https://semver.org/). Forked from claude-squad v1.0.20.

## Unreleased

### Added
- Pixel "BOOTHS" banner on the welcome screen (thin, widely spaced capitals; the two O's are
  booths, one lit), with a half-block version for smaller panes, and a one-line `❯ B[ ][▮]THS`
  logo in the booth list, help and narrow panes.
- `CLAUDE_BOOTHS_HOME` env var to keep config, state and worktrees somewhere other than `~/.claude-booths`.
- Brand: logo, design system (`BRAND.md`), website (`docs/`), demo GIF and promo loop,
  all reproducible from `docs/demo/`.
- Color themes: `claude-booths` (default, new brand palette), `monokai`, `vscode`, `ocean-green`, `dracula`, `nord`,
  `gruvbox`, `solarized`. Cycle with `t` in the app or use `claude-booths themes [name]`; saved in
  config as `theme`. Also styles each session's tmux status bar.

### Changed
- New welcome screen (booth logo, wordmark, keys card) replaces the big ASCII banner; it shrinks to fit
  small panes. Underlined tab bar, `▍booths  N open` list title, calmer menu. Default palette is now
  indigo primary + amber secondary.
- Preview scrolling: mouse wheel / shift+↑↓ move 3 lines (was 1, and the first step only
  entered scroll mode), a `SCROLL nn%` header shows you're looking at history and how to get
  back (Esc, or scroll past the bottom), and the menu shows the scroll key on every tab.
- Renamed to Claude Booths: binary `claude-booths`, config `~/.claude-booths`, tmux prefix
  `claudebooths_`. The UI calls sessions "booths".
- Branch picker lists branches checked out elsewhere greyed out and unselectable.

### Performance
- Pane previews are streamed from the session's tmux client instead of polling `capture-pane`
  (two idle sessions: 124 tmux processes per 10 s → 0).

### Fixed
- New booths got stuck on the agent's folder-trust screen, and a Shift+N prompt's Enter picked
  Claude Code's preselected "No, exit", so the booth died. Known trust and first-run screens
  (Claude Code, Gemini CLI, Codex trust; Aider's first-run questions) are now answered first.
- A prompt was typed into whatever screen the agent showed. Screens claude-booths won't answer
  for you (Codex's "Update now", unknown yes/no questions) now get an error in the UI and the
  prompt isn't sent.
- `-y` didn't work with Claude Code run through a wrapper or with flags, and missed its current
  permission prompt; with Aider it kept pressing Enter on an already answered question.
- `-p` was ignored by Shift+N: the new booth dialog offered (and started) the config's
  default program. With `-p`, that program is the only choice.
- Deleting a booth left its Terminal tab's tmux session running when the tab hadn't been
  opened in this run; leftover `claudebooths_term_*` sessions without a booth are removed at
  startup.
- Deleting a booth whose agent had already exited could fail and leave the booth; errors
  from confirmed actions are now shown instead of dropped, and killing a tmux session uses an
  exact name match (a prefix match could hit another booth's session).
- A booth whose tmux session died kept showing as ready; it now shows as paused (`r` starts
  the agent again in the same worktree).
- Deleting a booth no longer reattaches to every other booth's tmux session.
- Preview kept showing "Setting up workspace..." or another booth's screen when a booth's
  screen couldn't be read, so a ready booth looked stuck; it now says what's wrong.
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
- A long branch name ran into the diff stats in the list.
- Tests no longer touch the real config dir or the default tmux server.
