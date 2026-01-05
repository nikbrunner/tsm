# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a bash script that provides a tmux popup interface for quickly switching between sessions, creating new ones, and managing existing sessions using fzf. The script must run inside a tmux session.

## Core Architecture

### Main Entry Point

- Script starts at the bottom after function definitions
- Must run inside a tmux session (validates at line 8)

### Execution Flow

1. **Configuration** (lines 4-5): Layout settings via environment variables
2. **Dependency Check** (lines 24-28): Validates fzf is installed
3. **Session Listing** (lines 100-118): Gets sessions excluding current, prioritizes last session
4. **FZF Selection** (lines 120-141): Interactive picker with key bindings
5. **Action Handling** (lines 156-182): Processes user's choice (switch/create/kill)

### Key Functions

**`apply_layout()`** (lines 14-22)

- Applies tmux layout to new sessions
- Checks for executable layout script at `$TMUX_LAYOUTS_DIR/$TMUX_LAYOUT.sh`
- Called when creating new sessions

**`pick_window()`** (lines 51-98)

- Nested picker for windows within a session
- Supports switch, kill, and pane drilling via `Ctrl-O`

**`pick_pane()`** (lines 30-49)

- Deepest level picker for panes within a window
- Shows pane index and current command

### Key Features

**Last Session Priority**

- Retrieves last session from tmux option `@last_session` (line 113)
- Places it first in the list for quick switching

**Hierarchical Navigation**

- Sessions -> Windows -> Panes
- Each level accessible via `Ctrl-O`
- `Esc` returns to previous level

**Persistent Kill Mode**

- `Ctrl-X` kills and re-runs picker via `exec "$0"` (line 169)
- Allows killing multiple sessions without reopening

## Dependencies

**Required:**

- `tmux`: Must be running in a tmux session
- `fzf`: Interactive fuzzy finder

**Optional:**

- `gum`: Enhanced confirmations (not currently used in main flow)

## Environment Variables

- `TMUX_LAYOUT`: Layout name to apply on new sessions (default: `ide`)
- `TMUX_LAYOUTS_DIR`: Directory containing layout scripts (default: `~/.config/tmux/layouts`)

## Development Notes

### Testing the Script

Since this is tmux-specific:

1. Must run inside a tmux session
2. Test with `display-popup -E "./tsm"` for popup behavior
3. Create multiple sessions to test switching and killing

### Critical Behaviors

- Empty selections exit gracefully (line 152-154)
- New session names typed in fzf create sessions (lines 177-181)
- Kill action re-executes script to stay open (line 169)
