#!/bin/bash
# Builds a throwaway HOME with a sample repo for docs/demo/demo.tape. Sourced by the tape;
# never touches your real ~/.claude-booths or tmux server.
DEMO_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$DEMO_DIR/../.." && pwd)
export HOME=$(mktemp -d) TMUX_TMPDIR=$(mktemp -d) GIT_TERMINAL_PROMPT=0
unset TMUX
mkdir -p "$HOME/bin" "$HOME/.claude-booths"
(cd "$REPO_ROOT" && go build -o "$HOME/bin/claude-booths" .)
cp "$DEMO_DIR/agent" "$HOME/bin/agent"
export PATH="$HOME/bin:$PATH"
cat > "$HOME/.claude-booths/config.json" <<'JSON'
{"default_program":"agent","daemon_poll_interval":1000,"branch_prefix":"booths/","theme":"claude-booths"}
JSON
echo '{"help_screens_seen":255,"instances":[]}' > "$HOME/.claude-booths/state.json"
mkdir -p "$HOME/acme-shop/src" && cd "$HOME/acme-shop"
echo '# acme-shop' > README.md && echo 'export {}' > src/index.ts
git init -q -b main && git add -A && git -c user.name=demo -c user.email=demo@example.com commit -qm init
PS1='\[\e[38;2;139;124;255m\]~/acme-shop\[\e[0m\] \[\e[38;2;255;200;97m\]❯\[\e[0m\] '
