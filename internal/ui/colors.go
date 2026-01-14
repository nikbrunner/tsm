package ui

import "github.com/charmbracelet/lipgloss"

// ANSI colors (terminal-adaptive)
var (
	red         = lipgloss.Color("1")
	green       = lipgloss.Color("2")
	yellow      = lipgloss.Color("3")
	blue        = lipgloss.Color("4")
	white       = lipgloss.Color("7")
	brightBlack = lipgloss.Color("8")
	brightWhite = lipgloss.Color("15")
)

// Hex colors (terminal-independent)
var (
	hexClaudeOrange = lipgloss.Color("#DA7756")
	hexGitBlue      = lipgloss.Color("#61AFEF")
	hexGitGreen     = lipgloss.Color("#98C379")
	hexGitRed       = lipgloss.Color("#E06C75")
)

// FgColors defines all foreground (text) colors
type FgColors struct {
	Default  lipgloss.TerminalColor // Terminal default text
	Selected lipgloss.TerminalColor // Selected/highlighted items
	Muted    lipgloss.TerminalColor // De-emphasized text
	Accent   lipgloss.TerminalColor // Primary accent
	Subtle   lipgloss.TerminalColor // Secondary/subtle text
	Error    lipgloss.TerminalColor // Error text
	Border   lipgloss.TerminalColor // Border characters

	// Title bar
	TitleBar lipgloss.TerminalColor // Text on title bar

	// Table
	TableHeader lipgloss.TerminalColor // Column headers
	SessionName lipgloss.TerminalColor // Unselected session names
	WindowName  lipgloss.TerminalColor // Unselected window names

	// Claude status
	ClaudeHeader  lipgloss.TerminalColor // "CC" label
	ClaudeWorking lipgloss.TerminalColor // Spinner
	ClaudeWaiting lipgloss.TerminalColor // "?" icon
	ClaudeUrgent  lipgloss.TerminalColor // "!" icon

	// Git status
	GitFiles lipgloss.TerminalColor // File count
	GitAdd   lipgloss.TerminalColor // Additions
	GitDel   lipgloss.TerminalColor // Deletions
}

// BgColors defines all background colors
type BgColors struct {
	Default  lipgloss.TerminalColor // Terminal default (none)
	TitleBar lipgloss.TerminalColor // Title bar background
}

// Colors is the global color configuration
var Colors = struct {
	Fg FgColors
	Bg BgColors
}{
	Fg: FgColors{
		Default:  lipgloss.NoColor{},
		Selected: yellow,
		Muted:    brightBlack,
		Accent:   blue,
		Subtle:   white,
		Error:    red,
		Border:   brightBlack,

		TitleBar: brightWhite,

		TableHeader: lipgloss.NoColor{},
		SessionName: lipgloss.NoColor{},
		WindowName:  lipgloss.NoColor{},

		ClaudeHeader:  hexClaudeOrange,
		ClaudeWorking: yellow,
		ClaudeWaiting: green,
		ClaudeUrgent:  red,

		GitFiles: hexGitBlue,
		GitAdd:   hexGitGreen,
		GitDel:   hexGitRed,
	},
	Bg: BgColors{
		Default:  lipgloss.NoColor{},
		TitleBar: blue,
	},
}
