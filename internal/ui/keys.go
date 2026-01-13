package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application
type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	Expand        key.Binding
	Collapse      key.Binding
	Select        key.Binding
	Kill          key.Binding
	Create        key.Binding
	PickDirectory key.Binding
	CloneRepo     key.Binding
	Lazygit       key.Binding
	Bookmarks     key.Binding
	AddBookmark   key.Binding
	Quit          key.Binding
	Cancel        key.Binding
	Confirm       key.Binding
	Jump1         key.Binding
	Jump2         key.Binding
	Jump3         key.Binding
	Jump4         key.Binding
	Jump5         key.Binding
	Jump6         key.Binding
	Jump7         key.Binding
	Jump8         key.Binding
	Jump9         key.Binding
}

// DefaultKeyMap returns the default key bindings
// Navigation uses Ctrl+key or arrows, letters are reserved for filtering
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("ctrl+k", "up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("ctrl+j", "down"),
		key.WithHelp("↓", "down"),
	),
	Expand: key.NewBinding(
		key.WithKeys("ctrl+l", "right"),
		key.WithHelp("→", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("ctrl+h", "left"),
		key.WithHelp("←", "collapse"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "switch"),
	),
	Kill: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("C-x", "kill"),
	),
	Create: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("C-n", "new"),
	),
	PickDirectory: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("C-p", "projects"),
	),
	CloneRepo: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("C-r", "clone repo"),
	),
	Lazygit: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("C-g", "lazygit"),
	),
	Bookmarks: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("C-b", "bookmarks"),
	),
	AddBookmark: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("C-a", "add bookmark"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("C-c", "quit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("C-y", "confirm"),
	),
	Jump1: key.NewBinding(key.WithKeys("1")),
	Jump2: key.NewBinding(key.WithKeys("2")),
	Jump3: key.NewBinding(key.WithKeys("3")),
	Jump4: key.NewBinding(key.WithKeys("4")),
	Jump5: key.NewBinding(key.WithKeys("5")),
	Jump6: key.NewBinding(key.WithKeys("6")),
	Jump7: key.NewBinding(key.WithKeys("7")),
	Jump8: key.NewBinding(key.WithKeys("8")),
	Jump9: key.NewBinding(key.WithKeys("9")),
}

// helpItem formats a single help item (key + description)
func helpItem(key, desc string) string {
	return HelpKeyStyle.Render(key) + " " + HelpDescStyle.Render(desc)
}

// helpSep returns the separator between help items
func helpSep() string {
	return HelpSepStyle.Render(" · ")
}

// HelpNormal returns the help text for normal mode (two lines)
func HelpNormal() string {
	line1 := helpItem("type", "filter") + helpSep() +
		helpItem("C-j/k | ↑↓", "nav") + helpSep() +
		helpItem("C-h/l | ←→", "expand") + helpSep() +
		helpItem("C-x", "kill")
	line2 := helpItem("C-n", "new") + helpSep() +
		helpItem("C-p", "projects") + helpSep() +
		helpItem("C-b", "bookmarks") + helpSep() +
		helpItem("C-r", "clone") + helpSep() +
		helpItem("C-g", "lazygit")
	return line1 + "\n" + line2
}

// HelpFiltering returns the help text when filter is active
func HelpFiltering() string {
	return helpItem("esc", "clear") + helpSep() +
		helpItem("enter", "select") + helpSep() +
		helpItem("C-c", "quit")
}

// HelpConfirmKill returns the help text for kill confirmation mode
func HelpConfirmKill() string {
	return helpItem("C-x", "confirm") + helpSep() +
		helpItem("esc", "cancel")
}

// HelpCreate returns the help text for create mode
func HelpCreate() string {
	return helpItem("enter", "create") + helpSep() +
		helpItem("esc", "cancel")
}

// HelpPickDirectory returns the help text for directory picker mode
func HelpPickDirectory() string {
	return helpItem("↑↓", "nav") + helpSep() +
		helpItem("enter", "select") + helpSep() +
		helpItem("C-x", "remove") + helpSep() +
		helpItem("esc", "back")
}

// HelpConfirmRemoveFolder returns the help text for folder removal confirmation
func HelpConfirmRemoveFolder() string {
	return helpItem("C-x", "confirm") + helpSep() +
		helpItem("esc", "cancel")
}

// HelpCloneRepo returns the help text for clone repo mode
func HelpCloneRepo() string {
	return helpItem("↑↓", "nav") + helpSep() +
		helpItem("enter", "clone") + helpSep() +
		helpItem("esc", "back/cancel")
}

// HelpCloneRepoLoading returns the help text while loading repos
func HelpCloneRepoLoading() string {
	return helpItem("esc", "cancel")
}

// HelpCloneSuccess returns the help text after successful clone
func HelpCloneSuccess() string {
	return helpItem("enter", "switch to session") + helpSep() +
		helpItem("esc", "back to sessions")
}

// HelpBookmarks returns the help text for bookmarks mode
func HelpBookmarks() string {
	return helpItem("↑↓", "nav") + helpSep() +
		helpItem("enter", "open") + helpSep() +
		helpItem("C-p/n", "move") + helpSep() +
		helpItem("C-a", "add") + helpSep() +
		helpItem("C-x", "remove") + helpSep() +
		helpItem("esc", "back")
}
