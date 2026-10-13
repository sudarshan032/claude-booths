# Claude Style [![CI](https://github.com/sudarshan032/claude-style/actions/workflows/build.yml/badge.svg)](https://github.com/sudarshan032/claude-style/actions/workflows/build.yml) [![GitHub Release](https://img.shields.io/github/v/release/sudarshan032/claude-style)](https://github.com/sudarshan032/claude-style/releases/latest)

Claude Style is a terminal app that manages multiple [Claude Code](https://github.com/anthropics/claude-code), [Codex](https://github.com/openai/codex), [Gemini](https://github.com/google-gemini/gemini-cli) (and other local agents including [Aider](https://github.com/Aider-AI/aider)) in separate workspaces, allowing you to work on multiple tasks simultaneously.


![Claude Style Screenshot](assets/screenshot.png)

### Highlights
- Complete tasks in the background (including yolo / auto-accept mode!)
- Manage instances and tasks in one terminal window
- Review changes before applying them, checkout changes before pushing them
- Each task gets its own isolated git workspace, so no conflicts

### What's different from claude-squad

Claude Style is a fork of [smtg-ai/claude-squad](https://github.com/smtg-ai/claude-squad) with:

- **Streamed previews**: pane changes are pushed by tmux instead of polled with `capture-pane`
  ~12 times a second per session, so idle sessions spawn no processes at all.
- **No lost input while attached**: the TUI stops reading stdin during attach, so the first
  keystroke or the start of a paste is never swallowed.
- **Fetch never hangs**: background `git fetch` can't prompt for credentials on your TTY.
- **Branch picker knows what's in use**: branches checked out by another session (or your main
  folder) are shown greyed out and can't be picked, instead of the new session silently vanishing.
- **Prompt delivery waits for the agent to boot** (slow SessionStart hooks no longer eat the Enter).
- **Preview and Terminal tabs show the full pane** (no lost top/bottom row).
- Migrates `~/.claude-squad` config/state and running sessions on first run.

### Installation

```bash
curl -fsSL https://raw.githubusercontent.com/sudarshan032/claude-style/main/install.sh | bash
```

This puts the `claude-style` binary in `~/.local/bin`. To use a different name (e.g. `cs`):

```bash
curl -fsSL https://raw.githubusercontent.com/sudarshan032/claude-style/main/install.sh | bash -s -- --name cs
```

Or build from source (Go 1.23+): `go build -o ~/.local/bin/claude-style .`

### Prerequisites

- [tmux](https://github.com/tmux/tmux/wiki/Installing)
- [gh](https://cli.github.com/)

### Usage

```
Usage:
  claude-style [flags]
  claude-style [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  debug       Print debug information like config paths
  help        Help about any command
  reset       Reset all stored instances
  version     Print the version number of claude-style

Flags:
  -y, --autoyes          [experimental] If enabled, all instances will automatically accept prompts for claude code & aider
  -h, --help             help for claude-style
  -p, --program string   Program to run in new instances (e.g. 'aider --model ollama_chat/gemma3:1b')
```

Run the application with:

```bash
claude-style
```
NOTE: The default program is `claude` and we recommend using the latest version.

<br />

<b>Using Claude Style with other AI assistants:</b>
- For [Codex](https://github.com/openai/codex): Set your API key with `export OPENAI_API_KEY=<your_key>`
- Launch with specific assistants:
   - Codex: `claude-style -p "codex"`
   - Aider: `claude-style -p "aider ..."`
   - Gemini: `claude-style -p "gemini"`
- Make this the default, by modifying the config file (locate with `claude-style debug`)

<br />

#### Menu
The menu at the bottom of the screen shows available commands: 

##### Instance/Session Management
- `n` - Create a new session
- `N` - Create a new session with a prompt
- `D` - Kill (delete) the selected session
- `↑/j`, `↓/k` - Navigate between sessions

##### Actions
- `↵/o` - Attach to the selected session to reprompt
- `ctrl-q` - Detach from session
- `s` - Commit and push branch to github
- `c` - Checkout. Commits changes and pauses the session
- `r` - Resume a paused session
- `?` - Show help menu

##### Navigation
- `tab` - Switch between preview tab and diff tab
- `q` - Quit the application
- `shift-↓/↑` - scroll in diff view

### Configuration

Claude Style stores its configuration in `~/.claude-style/config.json`. You can find the exact path by running `claude-style debug`.

#### Profiles

Profiles let you define multiple named program configurations and switch between them when creating a new session. When more than one profile is defined, the session creation overlay shows a profile picker that you can navigate with `←`/`→`.

To configure profiles, add a `profiles` array to your config file and set `default_program` to the name of the profile to select by default:

```json
{
  "default_program": "claude",
  "profiles": [
    { "name": "claude", "program": "claude" },
    { "name": "codex", "program": "codex" },
    { "name": "aider", "program": "aider --model ollama_chat/gemma3:1b" }
  ]
}
```

Each profile has two fields:

| Field     | Description                                              |
|-----------|----------------------------------------------------------|
| `name`    | Display name shown in the profile picker                 |
| `program` | Shell command used to launch the agent for that profile  |

If no profiles are defined, Claude Style uses `default_program` directly as the launch command (the default is `claude`).

### FAQs

#### Failed to start new session

If you get an error like `failed to start new session: timed out waiting for tmux session`, update the
underlying program (ex. `claude`) to the latest version.

### How It Works

1. **tmux** to create isolated terminal sessions for each agent
2. **git worktrees** to isolate codebases so each session works on its own branch
3. A simple TUI interface for easy navigation and management

### License

[AGPL-3.0](LICENSE.md). Claude Style is a modified version of
[claude-squad](https://github.com/smtg-ai/claude-squad) by smtg-ai, also AGPL-3.0; the original
copyright and license are retained.

"Claude" is a trademark of Anthropic. This project is not affiliated with or endorsed by Anthropic.

