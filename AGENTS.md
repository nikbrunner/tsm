# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

helm ("Take the helm of your workspaces") is a Go TUI application for managing tmux sessions. It provides fuzzy filtering, session/window navigation, and Claude Code status integration. Built with Bubbletea/Lipgloss.

Part of the [Black Atom Industries](https://github.com/black-atom-industries) cockpit.

## Build Commands

```bash
make build          # Build binary to ./helm
make install        # Build and install to ~/.local/bin/helm (also installs git hooks)
make install-hooks  # Point git core.hooksPath at .githooks (pre-commit: fmt, lint, test)
make test           # Run tests
make fmt            # Format code
make lint           # Run golangci-lint
make tidy           # go mod tidy
```

For quick iteration: `go build -o helm ./cmd/helm/ && cp helm ~/.local/bin/helm`

## Architecture

```
cmd/helm/
  main.go                 # Entry point, subcommands (init, setup, bookmark, tmux-bindings, repos)
  repos.go                # repos subcommand (status, pull, push, dirty, rebuild)
  setup.go                # setup subcommand (bulk clone from ensure_cloned config)
internal/
  model/model.go          # Bubbletea Model - main state, Update/View logic
  ui/
    keys.go               # Key bindings (KeyMap) and help text
    styles.go             # Lipgloss styles (built from colors.go tokens)
    colors.go             # Semantic color tokens: Black Atom theme or ANSI-16 fallback
    columns.go            # Row rendering (sessions, windows, bookmarks)
    agentpanel.go         # AGENTS side panel (per-instance agent status)
    help.go               # ? keymap overlay
    scrolllist.go         # Generic scrollable list with filtering
    theme/                # Black Atom theme registry + generated themes (make themes)
  config/config.go        # YAML config (~/.config/black-atom/helm-tmux/config.yml)
  tmux/tmux.go            # tmux command wrappers (list, switch, kill)
  agent/
    status.go             # Agent (Claude Code, Pi) status file parsing
    liveness.go           # Process-tree liveness check behind status files
  git/
    status.go             # Git status per session (dirty, ahead/behind)
    repo.go               # Repo sync state (clean/dirty/ahead/behind/diverged)
  giturl/github.go         # Git URL parsing, clone, and GitHub API
hooks/helm-tmux-hook.sh   # Claude Code hook for status updates
```

### Bubbletea Model Flow

The model (`internal/model/model.go`) has these modes:

- **ModeNormal**: Session list with fuzzy filtering
- **ModeBookmarks**: Bookmarked repos (local dirs without active sessions)
- **ModePickDirectory**: Directory picker for new sessions
- **ModeCloneRepo**: Clone repos from GitHub
- **ModeCreate**: Text input for new session name
- **ModeConfirmKill**: Kill confirmation prompt
- **ModeConfirmRemoveFolder**: Folder removal confirmation
- **ModeHelp**: Full-keymap overlay (`?`)

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
- `Ctrl+r`: Open git remote in browser
- `Ctrl+d`: Download (clone) repo from GitHub
- `Ctrl+g`: Lazygit
- `1-9`: Jump to session (only when no filter active)
- `?`: Help overlay with full keymap (only when no filter active)
- Type letters: Fuzzy filter

The footer is a single compact hint bar (`ui.RenderHintBar`) derived from the
per-mode `Action` lists in `internal/ui/sidebar.go` — there are no button rows.

## Configuration

Config file: `~/.config/black-atom/helm-tmux/config.yml`

```yaml
theme: black-atom-jpn-koyo-yoru # Black Atom theme (empty = terminal ANSI colors)
layout: ide # Layout script for new sessions
layout_dir: ~/.config/tmux/layouts
claude_status_enabled: true # Show CC status indicator
cache_dir: ~/.cache/helm-tmux
dirty_walkthrough_command: "lazygit -p {}" # Command for 'helm repos dirty --walk'
```

Environment variables override config: `TMUX_LAYOUT`, `TMUX_LAYOUTS_DIR`, `TMUX_SESSION_PICKER_CLAUDE_STATUS=1`, `HELM_THEME`

## Theming (Black Atom adapter)

helm is a [Black Atom](https://github.com/black-atom-industries/core) adapter. The
`black-atom-adapter.json` manifest maps every collection to the Eta template
`internal/ui/theme/collection.template.go`, which renders one self-registering Go
file per theme (committed, so end users never need Deno).

- Set `theme:` in config (or `HELM_THEME`) to a theme key; the theme's own
  appearance wins over the `appearance` key. Empty = terminal-native ANSI-16 +
  reverse-video fallback.
- Regenerate after core changes: `make themes` (requires deno).
- The template is itself valid Go; the unrendered entry it registers at init is
  dropped by `Register`'s key guard (`internal/ui/theme/registry.go`).

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

## Agent Status Integration (Claude Code, Pi)

The hook (`hooks/helm-tmux-hook.sh`) writes JSON status files to `~/.cache/helm-tmux/<session>.<session_id>.status` — one file per agent instance, so multiple Claude instances in one tmux session don't overwrite each other (`{"state","ts","tool","session_id","transcript","cwd"}`; the legacy un-suffixed `state:timestamp` format still parses). The TUI polls these every second (`internal/agent`) and shows animated status indicators per session:

- `⠤⠆⠒⠰` (spinner) - Claude actively processing
- `?` - Claude waiting (0–5 min)
- `!` - Still waiting (5–15 min)
- `Z` - Idle (> 15 min)

With multiple instances, the list glyph shows the most active one (working > waiting > new). The **AGENTS side panel** (viewport ≥ 100×15, 35% of the content width via `ui.AgentPanelRatio`) lists every live instance of the selected session with state, elapsed time, and current/last tool; the footer counts them ("10 sessions · 3 agents"). Expanded window/pane rows carry a per-pane agent ident ("● claude" / "● pi") attributed via the process tree — reliable even when the pane command is just "node".

Because hooks don't fire on crash or SIGKILL, each poll also verifies via a process-tree check (`internal/agent/liveness.go`) that an agent process actually runs beneath the session's panes — stale status files are removed. Pi status works identically via `.pi-status` files. Liveness is per-kind-per-session: it cannot tell which of two same-kind instances died, so individual crashed instances age out via the stale thresholds instead.

> **Known limitation:** Claude Code's `Stop` hook has no `stop_reason` field to differentiate "idle/done" from "waiting for input." See upstream [anthropics/claude-code#13024](https://github.com/anthropics/claude-code/issues/13024).

---

## Project Tracking

Issues are tracked in [GitHub Issues](https://github.com/black-atom-industries/helm.tmux/issues) with the `helm` label.

Use the `gh` CLI to query and manage issues directly from Claude Code:

- `gh issue list` - Query issues
- `gh issue create` - Create new issues
- `gh issue edit` - Update status, labels, etc.

---

> **Note to Claude:** This file is named `AGENTS.md` with a symlink `CLAUDE.md -> AGENTS.md` because Anthropic's Claude Code does not yet support `AGENTS.md` as a context file. Once Claude Code supports `AGENTS.md` natively, the symlink can be removed.
