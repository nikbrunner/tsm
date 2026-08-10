# Helm for tmux

> Take the helm of your tmux workspaces. See also [helm.herdr](https://github.com/black-atom-industries/helm.herdr) for the Herdr version.

A TUI for managing tmux sessions — quick switching, fuzzy filtering, and workspace organization. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

Part of the [Black Atom Industries](https://github.com/black-atom-industries) cockpit — pairs with [radar.nvim](https://github.com/black-atom-industries/radar.nvim) for file navigation.

## Features

- Fuzzy filtering (just start typing)
- Ctrl-based navigation (`Ctrl+j/k`) to preserve filter input
- Number shortcuts for instant session switching (`1`-`9`)
- Expandable sessions to view windows
- Quick kill with confirmation (`Ctrl+x`)
- Create new sessions inline (`Ctrl+n`)
- Project picker (`Ctrl+p`)
- Bookmarks (`Ctrl+b`)
- Agent status integration (Claude Code, Pi): animated spinner per session,
  AGENTS side panel with per-instance state, elapsed time, and current tool
- Git status per session (dirty/ahead/behind)
- `?` help overlay with the full keymap

## Installation

### Prerequisites

- Go 1.21+
- tmux

### Build and Install

```sh
git clone https://github.com/black-atom-industries/helm.tmux.git
cd helm.tmux
make install
```

This builds the `helm` binary and installs it to `~/.local/bin/`.

## Setup

Add a key binding to your `~/.tmux.conf`:

```tmux
bind -n M-w display-popup -w90% -h80% -B -E "helm"
```

A lazygit-style large popup is recommended — the AGENTS side panel needs at
least a 100×15 viewport and hides on smaller popups (the rest of the UI works
at any size).

Reload your tmux configuration: `tmux source-file ~/.tmux.conf`

## Keybindings

| Key                   | Action                                  |
| --------------------- | --------------------------------------- |
| Type letters          | Fuzzy filter sessions                   |
| `Ctrl+j/k` or `↓`/`↑` | Navigate up/down                        |
| `Ctrl+h/l` or `←`/`→` | Collapse/Expand session windows         |
| `1`-`9`               | Jump to session (when no filter active) |
| `Enter`               | Switch to selected session/window       |
| `Ctrl+x`              | Kill with confirmation                  |
| `Ctrl+n`              | Create new session                      |
| `Ctrl+p`              | Project picker                          |
| `Ctrl+b`              | Bookmarks                               |
| `Ctrl+a`              | Add/remove bookmark                     |
| `Ctrl+r`              | Clone repo from GitHub                  |
| `Ctrl+g`              | Open lazygit                            |
| `?`                   | Help overlay (when no filter active)    |
| `q`/`Esc`             | Quit                                    |

The footer shows a compact hint bar with the current mode's actions; `?`
opens the full keymap.

## Configuration

Initialize config file:

```sh
helm init
```

Config location: `~/.config/black-atom/helm/config.yml`

## Repository Management

helm includes CLI subcommands for managing all repos under your configured `project_dirs`.

### Bulk Clone

```sh
helm setup
```

Clones all repositories listed in `ensure_cloned` config. Supports wildcards (`org/*`) via `gh` CLI and post-clone hooks.

### Sync Commands

```sh
helm repos status              # Show sync state of all repos
helm repos pull                # Fetch and pull (ff-only) clean repos
helm repos push                # Push all ahead repos (including dirty+ahead)
helm repos add <repo>          # Clone a repo into project_dirs (owner/repo or URL)
helm repos dirty               # Print paths of dirty repos
helm repos dirty --walk        # Run configured command on each dirty repo
helm repos rebuild             # Re-run post_clone hooks
```

All commands support `--json` for machine-readable output.

### Dirty Walkthrough

Configure a command to run on each dirty repo:

```yaml
# ~/.config/black-atom/helm/config.yml
dirty_walkthrough_command: "lazygit -p {}"
```

Then `helm repos dirty --walk` steps through each dirty repo with lazygit. Use `{}` as the path placeholder — works with any command.

## Claude Code Status Integration

Display Claude Code status for each session with an animated indicator.

### Setup

1. Copy the hook script:

   ```sh
   cp hooks/helm-hook.sh ~/.local/bin/
   chmod +x ~/.local/bin/helm-hook.sh
   ```

2. Add hooks to your `~/.claude/settings.json`:

   ```json
   {
     "hooks": {
       "SessionStart": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/helm-hook.sh SessionStart"
             }
           ]
         }
       ],
       "PreToolUse": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/helm-hook.sh PreToolUse"
             }
           ]
         }
       ],
       "Stop": [
         {
           "hooks": [
             { "type": "command", "command": "~/.local/bin/helm-hook.sh Stop" }
           ]
         }
       ],
       "Notification": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/helm-hook.sh Notification"
             }
           ]
         }
       ],
       "SessionEnd": [
         {
           "hooks": [
             {
               "type": "command",
               "command": "~/.local/bin/helm-hook.sh SessionEnd"
             }
           ]
         }
       ]
     }
   }
   ```

3. Enable in config (`~/.config/black-atom/helm/config.yml`):

   ```yaml
   claude_status_enabled: true
   ```

### Display

Sessions show Claude status as a single animated character:

- `⠤⠆⠒⠰` (spinner) - Claude actively processing
- `?` - Claude waiting for input
- `!` - Claude waiting for input > 5 minutes (needs attention)

## Pi Status Integration

Display Pi agent status for each session with an animated indicator.

### Setup

1. Copy both files to their destinations:

   ```sh
   # Hook script (called by the extension)
   cp hooks/helm-pi-hook.sh ~/.local/bin/
   chmod +x ~/.local/bin/helm-pi-hook.sh

   # Pi extension (auto-discovered)
   cp hooks/helm-pi-status.ts ~/.pi/agent/extensions/
   ```

2. Restart Pi (or use `/reload`)

3. Enable in config (`~/.config/black-atom/helm/config.yml`):

   ```yaml
   pi_status_enabled: true
   ```

Or via environment variable:

```sh
export TMUX_SESSION_PICKER_PI_STATUS=1
```

### Extension Events

The extension subscribes to these Pi events:

| Event              | Hook      | Status     |
| ------------------ | --------- | ---------- |
| `session_start`    | `start`   | New        |
| `agent_start`      | `working` | Processing |
| `agent_end`        | `waiting` | Idle       |
| `session_shutdown` | `end`     | Cleanup    |

### Display

Sessions show Pi status as a single animated character (same visual style as Claude Code):

- `⠤⠆⠒⠰` (spinner) - Pi actively processing
- `?` - Pi waiting for input
- `!` - Pi waiting for input > 5 minutes (needs attention)
- `Z` - Pi idle > 15 minutes

## Project Tracking

Issues are tracked in [GitHub Issues](https://github.com/black-atom-industries/helm.tmux/issues) with the `helm` label.

## License

MIT
