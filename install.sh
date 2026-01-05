#!/usr/bin/env bash
# Install script for tsm (tmux session manager)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Check for Go
if ! command -v go &>/dev/null; then
    echo "Error: Go is required to build tsm"
    echo "Install Go from https://go.dev/dl/"
    exit 1
fi

# Create ~/.local/bin if it doesn't exist
mkdir -p "$HOME/.local/bin"

# Build and install the binary
echo "Building tsm..."
cd "$SCRIPT_DIR"
go build -o tsm ./cmd/tsm/
cp tsm "$HOME/.local/bin/tsm"

# Install the Claude status hook
cp "$SCRIPT_DIR/hooks/tsm-hook.sh" "$HOME/.local/bin/tsm-hook"
chmod +x "$HOME/.local/bin/tsm-hook"

echo ""
echo "Installed:"
echo "  ~/.local/bin/tsm"
echo "  ~/.local/bin/tsm-hook"
echo ""
echo "Make sure ~/.local/bin is in your PATH."
echo ""
echo "Add to your ~/.tmux.conf:"
echo '  bind -n M-w display-popup -w65% -h50% -B -E "tsm"'
echo ""
echo "To enable Claude Code status integration (optional):"
echo ""
echo "1. Add hooks to ~/.claude/settings.json:"
cat << 'EOF'

{
  "hooks": {
    "PreToolUse": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tsm-hook PreToolUse"}]}
    ],
    "Stop": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tsm-hook Stop"}]}
    ],
    "Notification": [
      {"hooks": [{"type": "command", "command": "~/.local/bin/tsm-hook Notification"}]}
    ]
  }
}

EOF
echo "2. Enable in your ~/.tmux.conf:"
echo "   set-environment -g TMUX_SESSION_PICKER_CLAUDE_STATUS 1"
echo ""
echo "   Then reload: tmux source-file ~/.tmux.conf"
