# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

helm ("Take the helm of your workspaces") is a Go TUI application for managing tmux sessions. It provides fuzzy filtering, session/window navigation, and Claude Code status integration. Built with Bubbletea/Lipgloss.

Part of the [Black Atom Industries](https://github.com/black-atom-industries) cockpit.

## Build Commands

```bash
make build          # Build binary to ./helm
make install        # Build and install to ~/.local/bin/helm
make test           # Run tests
make fmt            # Format code
make lint           # Run golangci-lint
make tidy           # go mod tidy
```

For quick iteration: `go build -o helm ./cmd/helm/ && cp helm ~/.local/bin/helm`

## Architecture

```
cmd/helm/main.go          # Entry point, handles `helm init` subcommand
internal/
  model/model.go          # Bubbletea Model - main state, Update/View logic
  ui/
    keys.go               # Key bindings (KeyMap) and help text
    styles.go             # Lipgloss colors and styles
    columns.go            # Row rendering (sessions, windows, bookmarks)
    scrolllist.go         # Generic scrollable list with filtering
  config/config.go        # YAML config (~/.config/helm/config.yml)
  tmux/tmux.go            # tmux command wrappers (list, switch, kill)
  claude/status.go        # Claude Code status file parsing
  git/status.go           # Git status per session (dirty, ahead/behind)
  repos/config.go         # Repos base path config (~/.config/repos/)
  github/github.go        # GitHub API for repo listing
hooks/helm-hook.sh        # Claude Code hook for status updates
```

### Bubbletea Model Flow

The model (`internal/model/model.go`) has seven modes:
- **ModeNormal**: Session list with fuzzy filtering
- **ModeBookmarks**: Bookmarked repos (local dirs without active sessions)
- **ModePickDirectory**: Directory picker for new sessions
- **ModeCloneRepo**: Clone repos from GitHub
- **ModeCreate**: Text input for new session name
- **ModeConfirmKill**: Kill confirmation prompt
- **ModeConfirmRemoveFolder**: Folder removal confirmation

Key state:
- `sessions []tmux.Session` - Raw session data
- `items []Item` - Flattened view (sessions + expanded windows)
- `filter string` - Current filter text
- `cursor int` - Selected item index
- `projectList *ui.ScrollList[string]` - Directory picker state
- `cloneList *ui.ScrollList[string]` - Clone repo picker state

### Key Bindings

Navigation uses Ctrl modifiers to reserve letters for filtering:
- `Ctrl+j/k` or arrows: Navigate
- `Ctrl+h/l` or arrows: Collapse/Expand sessions
- `Ctrl+n`: Create new session
- `Ctrl+p`: Pick directory (projects)
- `Ctrl+b`: Bookmarks
- `Ctrl+x`: Kill (requires confirmation)
- `Ctrl+r`: Clone repo
- `Ctrl+g`: Lazygit
- `1-9`: Jump to session (only when no filter active)
- Type letters: Fuzzy filter

## Configuration

Config file: `~/.config/helm/config.yml`

```yaml
layout: ide                       # Layout script for new sessions
layout_dir: ~/.config/tmux/layouts
claude_status_enabled: true       # Show CC status indicator
cache_dir: ~/.cache/helm
```

Environment variables override config: `TMUX_LAYOUT`, `TMUX_LAYOUTS_DIR`, `TMUX_SESSION_PICKER_CLAUDE_STATUS=1`

## Testing

Must test inside tmux:
```bash
tmux display-popup -w50% -h35% -B -E "./helm"
```

### Automated Visual Testing

To test UI changes and capture a screenshot for visual verification:
```bash
tmux display-popup -w50% -h35% -B -E "~/.local/bin/helm" &
sleep 0.8
screencapture -x /tmp/helm_test.png
```

Then read `/tmp/helm_test.png` to visually verify the UI looks correct.

## Claude Status Integration

The hook (`hooks/helm-hook.sh`) writes status files to `~/.cache/helm/<session>.status`. The TUI reads these to show animated status indicators per session:
- `⠤⠆⠒⠰` (spinner) - Claude actively processing
- `?` - Claude waiting for input
- `!` - Claude waiting for input > 5 minutes

---

## Project Tracking

Issues are tracked in [Linear](https://linear.app/black-atom-industries) under the Development team with the `helm` label.

Use the Linear MCP to query and manage issues directly from Claude Code:
- `mcp__linear__list_issues` - Query issues
- `mcp__linear__create_issue` - Create new issues
- `mcp__linear__update_issue` - Update status, labels, etc.

---

> **Note to Claude:** This file is named `AGENTS.md` with a symlink `CLAUDE.md -> AGENTS.md` because Anthropic's Claude Code does not yet support `AGENTS.md` as a context file. Once Claude Code supports `AGENTS.md` natively, the symlink can be removed.
