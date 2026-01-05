#!/usr/bin/env bash
# Install script for tmux-session-picker

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Check for Go
if ! command -v go &>/dev/null; then
    echo "Error: Go is required to build tmux-session-picker"
    echo "Install Go from https://go.dev/dl/"
    exit 1
fi

# Create ~/.local/bin if it doesn't exist
mkdir -p "$HOME/.local/bin"

# Build and install the binary
echo "Building tmux-session-picker..."
cd "$SCRIPT_DIR"
go build -o tsp ./cmd/tsp/
cp tsp "$HOME/.local/bin/tsp"

# Install the Claude status hook
cp "$SCRIPT_DIR/hooks/claude-status-hook.sh" "$HOME/.local/bin/tmux-session-picker-hook"
chmod +x "$HOME/.local/bin/tmux-session-picker-hook"

echo ""
echo "Installed:"
echo "  ~/.local/bin/tsp"
echo "  ~/.local/bin/tmux-session-picker-hook"
echo ""
echo "Make sure ~/.local/bin is in your PATH."
echo ""
echo "Add to your ~/.tmux.conf:"
echo '  bind -n M-w display-popup -w65% -h50% -B -E "tsp"'
echo ""
echo "To enable Claude Code status integration (optional):"
echo ""
echo "1. Add hooks to ~/.claude/settings.json:"
cat << 'EOF'

{
  "hooks": {
    "PreToolUse": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tmux-session-picker-hook PreToolUse"}]}
    ],
    "Stop": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tmux-session-picker-hook Stop"}]}
    ],
    "Notification": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tmux-session-picker-hook Notification"}]}
    ]
  }
}

EOF
echo "2. Enable in your ~/.tmux.conf:"
echo "   set-environment -g TMUX_SESSION_PICKER_CLAUDE_STATUS 1"
echo ""
echo "   Then reload: tmux source-file ~/.tmux.conf"
