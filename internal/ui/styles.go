package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

	// Claude status styles
	ClaudeNewStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	ClaudeWorkingStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	ClaudeWaitingStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ClaudeWaitingUrgentStyle = lipgloss.NewStyle().
					Foreground(ColorError)

	// Git status styles (hardcoded hex colors - independent of terminal theme)
	GitFilesStyle = lipgloss.NewStyle().
			Foreground(HexGitFiles)

	GitAddStyle = lipgloss.NewStyle().
			Foreground(HexGitAdd)

	GitDelStyle = lipgloss.NewStyle().
			Foreground(HexGitDel)

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

	// Title bar style - inverted colors (colored background)
	TitleBarStyle = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(lipgloss.Color("15")). // bright white text on colored bg
			Bold(true)

	// Prompt style
	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Padding(0, 1)

	// State line style
	StateStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)

	// Table header style (subtle, dim)
	TableHeaderStyle = lipgloss.NewStyle().
				Padding(0, 1)

	// Table header text style (bold, dim)
	TableHeaderTextStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorDim)

	// CC header label style (bold, orange for Claude branding)
	CCHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(HexClaudeOrange)
)

// RenderBorder returns a horizontal border line
func RenderBorder(width int) string {
	return BorderStyle.Render(strings.Repeat("─", width))
}

// RenderDottedBorder returns a subtle dotted horizontal line
func RenderDottedBorder(width int) string {
	return BorderStyle.Render(strings.Repeat("·", width))
}

// RenderTitleBar renders the inverted title bar with logo on left and view name on right
func RenderTitleBar(logo, viewName string, width int) string {
	// Account for padding in AppStyle (1 on each side)
	innerWidth := width - AppBorderOverheadX
	if innerWidth < 10 {
		innerWidth = 40 // fallback for initial render
	}

	// Calculate spacing between logo and view name
	spacing := innerWidth - len(logo) - len(viewName) - 2 // -2 for padding spaces
	if spacing < 1 {
		spacing = 1
	}

	content := " " + logo + strings.Repeat(" ", spacing) + viewName + " "
	return TitleBarStyle.Width(innerWidth).Render(content)
}

// RenderPrompt renders the prompt line with optional filter text
func RenderPrompt(filter string, width int) string {
	innerWidth := width - AppBorderOverheadX
	if innerWidth < 10 {
		innerWidth = 40 // fallback for initial render
	}
	prompt := "> " + filter
	return PromptStyle.Width(innerWidth).Render(prompt)
}

// RenderFooter renders the 3-line footer (notification, state, hints)
func RenderFooter(notification, state, hints string, isError bool, width int) string {
	innerWidth := width - AppBorderOverheadX
	if innerWidth < 10 {
		innerWidth = 40 // fallback for initial render
	}
	var b strings.Builder

	// Border
	b.WriteString(RenderBorder(innerWidth))
	b.WriteString("\n")

	// Notification line (always 1 line, even if empty)
	if notification != "" {
		if isError {
			b.WriteString(ErrorMessageStyle.Width(innerWidth).Render(notification))
		} else {
			b.WriteString(MessageStyle.Width(innerWidth).Render(notification))
		}
	} else {
		b.WriteString(strings.Repeat(" ", innerWidth))
	}
	b.WriteString("\n")

	// State line (always 1 line, even if empty)
	if state != "" {
		b.WriteString(StateStyle.Width(innerWidth).Render(state))
	} else {
		b.WriteString(strings.Repeat(" ", innerWidth))
	}
	b.WriteString("\n")

	// Hints line
	b.WriteString(FooterStyle.Width(innerWidth).Render(hints))

	return b.String()
}

// ClaudeSpinnerFrames is the 4-frame braille spinner for "working" state
// Uses bottom 4 dots (positions 2,3,5,6) for better vertical alignment
var ClaudeSpinnerFrames = []string{"⠤", "⠆", "⠒", "⠰"}

// ClaudeWaitThreshold is the duration after which "waiting" escalates from ? to !
const ClaudeWaitThreshold = 5 * time.Minute

// FormatClaudeIcon formats the Claude status as a single character icon
// animationFrame cycles 0-3 for the spinner, waitDuration determines ? vs !
func FormatClaudeIcon(state string, animationFrame int, waitDuration time.Duration) string {
	switch state {
	case "new":
		// Don't show icon for "new" - it's just noise
		return " "
	case "working":
		// Animated spinner
		frame := animationFrame % len(ClaudeSpinnerFrames)
		return ClaudeWorkingStyle.Render(ClaudeSpinnerFrames[frame])
	case "waiting":
		// Escalate from ? to ! after threshold
		if waitDuration >= ClaudeWaitThreshold {
			return ClaudeWaitingUrgentStyle.Render("!")
		}
		return ClaudeWaitingStyle.Render("?")
	default:
		return " "
	}
}

// GitStatusColumnWidth is the fixed width for the git status column
const GitStatusColumnWidth = 20 // fits "99 files +99 -99"

// FormatGitStatus formats git status for display
// Returns empty string for clean repos (no indicator shown)
// Format: 3 files +44 -7 (files blue, +additions green, -deletions red)
func FormatGitStatus(dirty, additions, deletions int) string {
	if dirty == 0 && additions == 0 && deletions == 0 {
		return ""
	}

	var parts []string

	if dirty > 0 {
		label := "files"
		if dirty == 1 {
			label = "file"
		}
		parts = append(parts, GitFilesStyle.Render(fmt.Sprintf("%d %s", dirty, label)))
	}
	if additions > 0 {
		parts = append(parts, GitAddStyle.Render(fmt.Sprintf("+%d", additions)))
	}
	if deletions > 0 {
		parts = append(parts, GitDelStyle.Render(fmt.Sprintf("-%d", deletions)))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// GitStatusWidth returns the visual width of a git status string (without ANSI codes)
func GitStatusWidth(dirty, additions, deletions int) int {
	if dirty == 0 && additions == 0 && deletions == 0 {
		return 0
	}

	var parts []string

	if dirty > 0 {
		label := "files"
		if dirty == 1 {
			label = "file"
		}
		parts = append(parts, fmt.Sprintf("%d %s", dirty, label))
	}
	if additions > 0 {
		parts = append(parts, fmt.Sprintf("+%d", additions))
	}
	if deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d", deletions))
	}

	if len(parts) == 0 {
		return 0
	}

	return len(strings.Join(parts, " "))
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
