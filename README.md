# tmux-session-picker

A TUI for quickly switching between tmux sessions, creating new ones, and managing existing sessions.

Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for a fast, responsive interface with vim-style navigation.

## Features

- Vim-style navigation (`j`/`k`, `h`/`l`)
- Number shortcuts for instant session switching (`1`-`9`)
- Expandable sessions to view windows
- Quick kill with confirmation (`x`) or instant double-tap (`xx`)
- Create new sessions inline
- Claude Code status integration
- Last session indicator (󰒮)

## Installation

### Prerequisites

- Go 1.21+
- tmux

### Build and Install

```sh
git clone https://github.com/nikbrunner/tmux-session-picker.git
cd tmux-session-picker
make install
```

This builds the `tsp` binary and installs it to `~/.local/bin/`.

## Setup

Add a key binding to your `~/.tmux.conf`:

```tmux
bind -n M-w display-popup -w50% -h35% -B -E "tsp"
```

Reload your tmux configuration: `tmux source-file ~/.tmux.conf`

## Keybindings

| Key | Action |
|-----|--------|
| `j`/`k` or `↓`/`↑` | Navigate up/down |
| `h`/`l` or `←`/`→` | Collapse/Expand session windows |
| `1`-`9` | Jump to session (or window when expanded) |
| `Enter` | Switch to selected session/window |
| `x` | Kill with confirmation |
| `xx` | Instant kill (double-tap) |
| `c` | Create new session |
| `q`/`Esc` | Quit |

## Claude Code Status Integration

Optionally display Claude Code status for each session.

### Setup

1. Install the hook script (included with `make install`)

2. Add hooks to your `~/.claude/settings.json`:

   ```json
   {
     "hooks": {
       "PreToolUse": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/tmux-session-picker-hook PreToolUse" }] }],
       "Stop": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/tmux-session-picker-hook Stop" }] }],
       "Notification": [{ "hooks": [{ "type": "command", "command": "~/.local/bin/tmux-session-picker-hook Notification" }] }]
     }
   }
   ```

3. Enable in `~/.tmux.conf`:

   ```tmux
   set-environment -g TMUX_SESSION_PICKER_CLAUDE_STATUS 1
   ```

### Display

Sessions show Claude status:
- `[CC: new]` - New Claude session (dim)
- `[CC: working]` - Claude actively processing (yellow)
- `[CC: waiting]` - Claude finished, waiting for input (green)

## Layout Support

Apply layouts to new sessions via environment variables:

```bash
export TMUX_LAYOUT="ide"
export TMUX_LAYOUTS_DIR="$HOME/.config/tmux/layouts"
```

## License

MIT
