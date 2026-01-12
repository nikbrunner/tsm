package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ANSI 16 colors - adapts to terminal theme
// 0-7: black, red, green, yellow, blue, magenta, cyan, white
// 8-15: bright variants
var (
	ColorPrimary   = lipgloss.Color("4") // Blue
	ColorSecondary = lipgloss.Color("7") // White/light gray
	ColorSuccess   = lipgloss.Color("2") // Green
	ColorWarning   = lipgloss.Color("3") // Yellow
	ColorError     = lipgloss.Color("1") // Red
	ColorDim       = lipgloss.Color("8") // Bright black (dark gray)
	ColorClaude    = lipgloss.Color("5") // Magenta (distinctive for Claude)
)

// Border and padding overhead for the app container
const (
	// AppBorderOverhead is the total cells used by border + padding per axis
	AppBorderOverheadX = 4 // left border + left padding + right padding + right border
	AppBorderOverheadY = 2 // top border + bottom border (no vertical padding)
)

// Styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorDim).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(0, 1)

	MessageStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Padding(0, 1)

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Padding(0, 1)

	// Session row styles
	SessionStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SessionSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true)

	// Window row styles (indented)
	WindowStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(10)

	WindowSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				PaddingLeft(10).
				Bold(true)

	// Text styles
	IndexStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Width(3)

	IndexSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Width(3)

	SessionNameSelectedStyle = lipgloss.NewStyle().
					Foreground(ColorWarning).
					Bold(true)

	WindowNameSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	ExpandedIcon          = lipgloss.NewStyle().Foreground(ColorPrimary).Render("▼")
	ExpandedIconSelected  = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render("▼")
	CollapsedIcon         = lipgloss.NewStyle().Foreground(ColorDim).Render("▶")
	CollapsedIconSelected = lipgloss.NewStyle().Foreground(ColorDim).Bold(true).Render("▶")

	TimeStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	TimeSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Bold(true)

	LastIcon         = lipgloss.NewStyle().Foreground(ColorWarning).Render("󰒮")
	LastIconSelected = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render("󰒮")

	// Claude status styles
	ClaudeNewStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	ClaudeWorkingStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	ClaudeWaitingStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ClaudeLabelStyle = lipgloss.NewStyle().
				Foreground(ColorClaude)

	// Git status styles
	GitDirtyStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)

	GitAheadStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	GitBehindStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	// Input styles
	InputPromptStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Filter style
	FilterStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Statusline style
	StatuslineStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)
)

// RenderBorder returns a horizontal border line
func RenderBorder(width int) string {
	return BorderStyle.Render(strings.Repeat("─", width))
}

// FormatClaudeStatus formats the Claude status for display
// animationFrame cycles 0-2 for animated states
func FormatClaudeStatus(state string, animationFrame int) string {
	if state == "" {
		return ""
	}

	label := ClaudeLabelStyle.Render("CC:")

	switch state {
	case "new":
		// Don't show badge for "new" - it's just noise
		return ""
	case "working":
		// Animated ellipses: .  ..  ...
		dots := []string{".  ", ".. ", "..."}
		return "[" + label + " " + ClaudeWorkingStyle.Render(dots[animationFrame]) + "]"
	case "waiting":
		// Prominent - needs user attention
		return "[" + label + " " + ClaudeWorkingStyle.Render("?") + "]"
	default:
		return ""
	}
}

// GitStatusColumnWidth is the fixed width for the git status column
const GitStatusColumnWidth = 16 // fits [99cf +99 -99]

// FormatGitStatus formats git status for display
// Returns empty string for clean repos (no indicator shown)
// Format: [13cf +2 -1] (compact: cf=changed files, +=ahead, -=behind)
func FormatGitStatus(dirty, ahead, behind int) string {
	if dirty == 0 && ahead == 0 && behind == 0 {
		return ""
	}

	var parts []string

	if dirty > 0 {
		parts = append(parts, GitDirtyStyle.Render(fmt.Sprintf("%dcf", dirty)))
	}
	if ahead > 0 {
		parts = append(parts, GitAheadStyle.Render(fmt.Sprintf("+%d", ahead)))
	}
	if behind > 0 {
		parts = append(parts, GitBehindStyle.Render(fmt.Sprintf("-%d", behind)))
	}

	if len(parts) == 0 {
		return ""
	}

	return "[" + strings.Join(parts, " ") + "]"
}

// GitStatusWidth returns the visual width of a git status string (without ANSI codes)
func GitStatusWidth(dirty, ahead, behind int) int {
	if dirty == 0 && ahead == 0 && behind == 0 {
		return 0
	}

	var parts []string

	if dirty > 0 {
		parts = append(parts, fmt.Sprintf("%dcf", dirty))
	}
	if ahead > 0 {
		parts = append(parts, fmt.Sprintf("+%d", ahead))
	}
	if behind > 0 {
		parts = append(parts, fmt.Sprintf("-%d", behind))
	}

	if len(parts) == 0 {
		return 0
	}

	// [parts joined by " "]
	return len("[") + len(strings.Join(parts, " ")) + len("]")
}

// ScrollbarChars returns scrollbar characters for each visible line
// totalItems: total number of items in the list
// visibleItems: number of items currently visible
// scrollOffset: current scroll position (first visible item index)
// height: number of lines to render scrollbar for
func ScrollbarChars(totalItems, visibleItems, scrollOffset, height int) []string {
	result := make([]string, height)

	// No scrollbar needed if all items fit
	if totalItems <= visibleItems || height <= 0 {
		for i := range result {
			result[i] = " "
		}
		return result
	}

	// Calculate thumb size (minimum 1 line)
	thumbSize := (visibleItems * height) / totalItems
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position
	scrollRange := totalItems - visibleItems
	trackRange := height - thumbSize
	thumbPos := 0
	if scrollRange > 0 && trackRange > 0 {
		thumbPos = (scrollOffset * trackRange) / scrollRange
	}

	// Build scrollbar
	trackChar := BorderStyle.Render("│")
	thumbChar := lipgloss.NewStyle().Foreground(ColorSecondary).Render("┃")

	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = thumbChar
		} else {
			result[i] = trackChar
		}
	}

	return result
}
