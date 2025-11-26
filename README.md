# tmux-session-picker

Pop up fzf in tmux to quickly switch between sessions, create new ones, or manage existing sessions.

## Installation

### Option 1: Direct Installation

1.  Download the script:
    ```sh
    curl -O https://raw.githubusercontent.com/nikbrunner/tmux-session-picker/main/tmux-session-picker
    ```
2.  Make it executable:
    ```sh
    chmod +x tmux-session-picker
    ```
3.  Move it to a directory in your `PATH`:
    ```sh
    mv tmux-session-picker ~/.local/bin/
    ```

### Option 2: Development Installation (via symlink)

If you're forking or modifying the script, use a symlink so changes are immediately available:

1.  Clone the repository:
    ```sh
    git clone https://github.com/nikbrunner/tmux-session-picker.git
    cd tmux-session-picker
    ```
2.  Make it executable:
    ```sh
    chmod +x tmux-session-picker
    ```
3.  Create a symlink in your `PATH`:
    ```sh
    ln -s "$(pwd)/tmux-session-picker" ~/.local/bin/tmux-session-picker
    ```

## Dependencies

- [fzf](https://github.com/junegunn/fzf): For the fuzzy-finding interface.
- [gum](https://github.com/charmbracelet/gum) (optional): For enhanced confirmations.

## Setup

Add a key binding to your `~/.tmux.conf` to launch the session picker:

```tmux
# ~/.tmux.conf
bind -n M-w display-popup -w65% -h35% -B -E "tmux-session-picker"
```

Reload your tmux configuration: `tmux source-file ~/.tmux.conf`

## Usage

1. Press your configured key binding (e.g., `Alt-w`).
2. An fzf window will pop up listing all sessions (except the current one).
3. Sessions are sorted by recency (most recently used first) and show relative time (e.g., "5m ago", "2h ago") to help identify stale sessions.

### Keybindings inside the picker

- `Enter`: Switch to the selected session, or create a new session if you typed a name
- `Ctrl-O`: Open window picker for the selected session
- `Ctrl-X`: Kill the selected session (picker stays open for more actions)
- `Esc`: Cancel and close the picker

### Window and Pane Navigation

When you press `Ctrl-O` on a session, you get a window picker with:

- `Enter`: Switch to the selected window
- `Ctrl-O`: Open pane picker for the selected window
- `Ctrl-X`: Kill the selected window
- `Esc`: Go back to session picker

### Creating New Sessions

Simply type a name that doesn't match any existing session and press `Enter` to create and switch to a new session.

## Layout Support

The picker supports automatic layout application for new sessions via environment variables:

```bash
# Set default layout (default: ide)
export TMUX_LAYOUT="ide"

# Set layouts directory (default: ~/.config/tmux/layouts)
export TMUX_LAYOUTS_DIR="$HOME/.config/tmux/layouts"

# Disable layouts
export TMUX_LAYOUT=""
```

Layout scripts receive the session name and working directory as arguments.

## License

MIT
