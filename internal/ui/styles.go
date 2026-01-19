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

	// ScrollbarColumnWidth is the space used by scrollbar + separator
	ScrollbarColumnWidth = 2 // scrollbar char + space
)

// Styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(Colors.Fg.Border).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Colors.Fg.Accent).
			Padding(0, 1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Subtle).
			Padding(0, 1)

	MessageStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Accent).
			Padding(0, 1)

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.Error).
				Padding(0, 1)

	// Session row styles
	SessionStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SessionSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true).
				Background(Colors.Bg.Selected)

	// Window row styles (indented)
	WindowStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(10)

	WindowSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				PaddingLeft(10).
				Bold(true).
				Background(Colors.Bg.Selected)

	// Pane row styles (further indented)
	PaneStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(14)

	PaneSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				PaddingLeft(14).
				Bold(true).
				Background(Colors.Bg.Selected)

	// Text styles
	IndexStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Subtle).
			Width(3)

	IndexSelectedStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.Selected).
				Background(Colors.Bg.Selected).
				Bold(true).
				Width(3)

	SessionNameStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.SessionName)

	SessionNameSelectedStyle = lipgloss.NewStyle().
					Foreground(Colors.Fg.Selected).
					Background(Colors.Bg.Selected).
					Bold(true)

	WindowNameStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.WindowName)

	WindowNameSelectedStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.Selected).
				Background(Colors.Bg.Selected).
				Bold(true)

	ExpandedIcon          = lipgloss.NewStyle().Foreground(Colors.Fg.Accent).Render("▼")
	ExpandedIconSelected  = lipgloss.NewStyle().Foreground(Colors.Fg.Accent).Background(Colors.Bg.Selected).Bold(true).Render("▼")
	CollapsedIcon         = lipgloss.NewStyle().Foreground(Colors.Fg.Muted).Render("▶")
	CollapsedIconSelected = lipgloss.NewStyle().Foreground(Colors.Fg.Muted).Background(Colors.Bg.Selected).Bold(true).Render("▶")

	TimeStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted)

	TimeSelectedStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.Muted).
				Background(Colors.Bg.Selected).
				Bold(true)

	// Claude status styles
	ClaudeNewStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted)

	ClaudeWorkingStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.ClaudeWorking)

	ClaudeWaitingStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.ClaudeWaiting)

	ClaudeWaitingUrgentStyle = lipgloss.NewStyle().
					Foreground(Colors.Fg.ClaudeUrgent)

	// Git status styles
	GitFilesStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.GitFiles)

	GitAddStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.GitAdd)

	GitDelStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.GitDel)

	GitLoadingStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted)

	// Input styles
	InputPromptStyle = lipgloss.NewStyle().
				Foreground(Colors.Fg.Accent)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Accent).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted)

	// Filter style
	FilterStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Selected).
			Bold(true)

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Border)

	// Statusline style
	StatuslineStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted).
			Padding(0, 1)

	// Title bar style - inverted colors (colored background)
	TitleBarStyle = lipgloss.NewStyle().
			Background(Colors.Bg.TitleBar).
			Foreground(Colors.Fg.TitleBar).
			Bold(true)

	// Prompt style
	PromptStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Accent).
			Padding(0, 1)

	// State line style
	StateStyle = lipgloss.NewStyle().
			Foreground(Colors.Fg.Muted).
			Padding(0, 1)

	// Table header style (subtle, dim)
	TableHeaderStyle = lipgloss.NewStyle().
				Padding(0, 1)

	// Table header text style (bold)
	TableHeaderTextStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Colors.Fg.TableHeader)

	// CC header label style (bold, orange for Claude branding)
	CCHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Colors.Fg.ClaudeHeader)
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
func FormatGitStatus(dirty, additions, deletions int, selected bool) string {
	if dirty == 0 && additions == 0 && deletions == 0 {
		return ""
	}

	// Apply background when selected
	filesStyle := GitFilesStyle
	addStyle := GitAddStyle
	delStyle := GitDelStyle
	if selected {
		filesStyle = filesStyle.Background(Colors.Bg.Selected)
		addStyle = addStyle.Background(Colors.Bg.Selected)
		delStyle = delStyle.Background(Colors.Bg.Selected)
	}

	var parts []string

	if dirty > 0 {
		label := "files"
		if dirty == 1 {
			label = "file"
		}
		parts = append(parts, filesStyle.Render(fmt.Sprintf("%d %s", dirty, label)))
	}
	if additions > 0 {
		parts = append(parts, addStyle.Render(fmt.Sprintf("+%d", additions)))
	}
	if deletions > 0 {
		parts = append(parts, delStyle.Render(fmt.Sprintf("-%d", deletions)))
	}

	if len(parts) == 0 {
		return ""
	}

	// Join with styled spaces when selected
	if selected {
		spacer := lipgloss.NewStyle().Background(Colors.Bg.Selected).Render(" ")
		return strings.Join(parts, spacer)
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
	thumbChar := lipgloss.NewStyle().Foreground(Colors.Fg.Subtle).Render("┃")

	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = thumbChar
		} else {
			result[i] = trackChar
		}
	}

	return result
}
