#!/usr/bin/env bash
# Pi status hook - writes status to ~/.cache/helm/<session>.pi-status
# Used by helm to display Pi status per session
#
# This script is called by the Pi extension (helm-pi-status.ts).
# See that file for installation instructions.

STATUS_DIR="$HOME/.cache/helm-tmux"
mkdir -p "$STATUS_DIR"

# Get tmux session name
TMUX_SESSION=$(tmux display-message -p '#{session_name}' 2>/dev/null)
[[ -z "$TMUX_SESSION" ]] && exit 0

STATUS_FILE="$STATUS_DIR/${TMUX_SESSION}.pi-status"
TIMESTAMP=$(date +%s)

# Event from Pi extension
EVENT="$1"

case "$EVENT" in
    "start")
        echo "new:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "working")
        echo "working:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "waiting")
        echo "waiting:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "end")
        rm -f "$STATUS_FILE"
        ;;
esac

exit 0