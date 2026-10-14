#!/bin/bash
# Throwaway state + sample repo for docs/demo/demo.tape. Sourced by the tape. Uses your real
# Claude Code login (HOME is untouched) but never your real ~/.claude-booths or tmux server.
DEMO_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$DEMO_DIR/../.." && pwd)
export CLAUDE_BOOTHS_HOME=$(mktemp -d) TMUX_TMPDIR=$(mktemp -d) GIT_TERMINAL_PROMPT=0
unset TMUX CLAUDE_CODE_CHILD_SESSION
mkdir -p "$CLAUDE_BOOTHS_HOME/bin"
(cd "$REPO_ROOT" && go build -o "$CLAUDE_BOOTHS_HOME/bin/claude-booths" .)
# The agent: Claude Code with your hooks/plugins off so the recording shows stock Claude Code.
# DEMO_CLAUDE must skip the folder-trust prompt (e.g. a wrapper that marks the worktree trusted).
cat > "$CLAUDE_BOOTHS_HOME/bin/claude-demo" <<SH
#!/bin/bash
exec ${DEMO_CLAUDE:-claude --dangerously-skip-permissions} --settings '{"disableAllHooks":true}'
SH
chmod +x "$CLAUDE_BOOTHS_HOME/bin/claude-demo"
export PATH="$CLAUDE_BOOTHS_HOME/bin:$PATH"
cat > "$CLAUDE_BOOTHS_HOME/config.json" <<JSON
{"default_program":"$CLAUDE_BOOTHS_HOME/bin/claude-demo","daemon_poll_interval":1000,"branch_prefix":"booths/","theme":"claude-booths"}
JSON
echo '{"help_screens_seen":255,"instances":[]}' > "$CLAUDE_BOOTHS_HOME/state.json"
cp -R "$DEMO_DIR/acme-shop" "$CLAUDE_BOOTHS_HOME/acme-shop" && cd "$CLAUDE_BOOTHS_HOME/acme-shop"
git init -q -b main && git add -A && git -c user.name=demo -c user.email=demo@example.com commit -qm "acme-shop"
PS1='\[\e[38;2;139;124;255m\]~/acme-shop\[\e[0m\] \[\e[38;2;245;165;36m\]❯\[\e[0m\] '
