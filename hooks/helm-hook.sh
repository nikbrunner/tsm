#!/usr/bin/env bash
# Claude Code hook - writes status to ~/.cache/helm/
# Used by helm to display Claude status per session
#
# Status file format is JSON: {"state":"working","ts":1730000000,"tool":"Bash",...}
# helm also still parses the legacy "state:timestamp" format.

STATUS_DIR="$HOME/.cache/helm-tmux"
mkdir -p "$STATUS_DIR"

# Read JSON from stdin (required by Claude Code hooks)
INPUT=$(cat)

# Get the tmux session THIS process runs in. $TMUX_PANE targets the hook's
# own pane — display-message without a target would report the currently
# focused client's session instead, misattributing the status whenever the
# user is looking at a different session while the hook fires.
if [[ -n "$TMUX_PANE" ]]; then
    TMUX_SESSION=$(tmux display-message -t "$TMUX_PANE" -p '#{session_name}' 2>/dev/null)
else
    TMUX_SESSION=$(tmux display-message -p '#{session_name}' 2>/dev/null)
fi
[[ -z "$TMUX_SESSION" ]] && exit 0

HOOK_TYPE="$1"
TIMESTAMP=$(date +%s)

# One status file per agent instance: <session>.<session_id>.status, so two
# Claude instances in the same tmux session don't overwrite each other.
# Falls back to the legacy <session>.status when no session_id is available.
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty' 2>/dev/null)
if [[ -n "$SESSION_ID" ]]; then
    STATUS_FILE="$STATUS_DIR/${TMUX_SESSION}.${SESSION_ID}.status"
else
    STATUS_FILE="$STATUS_DIR/${TMUX_SESSION}.status"
fi

# write_status <state> — persist state plus context from the hook payload
# (tool name, session id, transcript path). Falls back to a minimal JSON
# object if jq is unavailable or the payload is malformed.
write_status() {
    local state="$1"
    echo "$INPUT" | jq -c --arg state "$state" --argjson ts "$TIMESTAMP" \
        '{state: $state, ts: $ts, tool: (.tool_name // ""), session_id: (.session_id // ""), transcript: (.transcript_path // ""), cwd: (.cwd // "")}' \
        >"$STATUS_FILE" 2>/dev/null ||
        echo "{\"state\":\"$state\",\"ts\":$TIMESTAMP}" >"$STATUS_FILE"
}

case "$HOOK_TYPE" in
    "SessionStart")
        write_status "new"
        ;;
    "PreToolUse")
        # AskUserQuestion means Claude is asking the user something — treat as waiting
        TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null)
        if [[ "$TOOL_NAME" == "AskUserQuestion" ]]; then
            write_status "waiting"
        else
            write_status "working"
        fi
        ;;
    "Stop"|"SubagentStop"|"Notification")
        # Running background tasks keep the session logically working
        BG_COUNT=$(echo "$INPUT" | jq -r '(.background_tasks // []) | length' 2>/dev/null)
        if [[ "$BG_COUNT" =~ ^[0-9]+$ ]] && ((BG_COUNT > 0)); then
            write_status "working"
        else
            write_status "waiting"
        fi
        ;;
    "SessionEnd")
        # Clean up status file when Claude session ends
        rm -f "$STATUS_FILE"
        ;;
esac

exit 0
