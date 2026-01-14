package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/black-atom-industries/helm/internal/claude"
	"github.com/black-atom-industries/helm/internal/git"
)

// RowLayout holds calculated column widths for consistent alignment across rows
type RowLayout struct {
	NameWidth      int
	GitStatusWidth int
}

// RowOpts contains options for rendering a generic row
// Required: Num, Name, Selected
// Optional: all others (use zero values to omit)
type RowOpts struct {
	// Required
	Num      int
	Name     string
	Selected bool

	// Optional - set to enable
	ShowExpandIcon bool           // Show ▸/▾ expand indicator
	Expanded       bool           // Expansion state
	LastActivity   *time.Time     // Show time ago if set
	GitStatus      *git.Status    // Show git status if set
	ClaudeStatus   *claude.Status // Show claude status if set
	AnimFrame      int            // Animation frame for claude status
}

// WindowRowOpts contains per-row options for rendering a window
type WindowRowOpts struct {
	Selected bool
}

// PaneRowOpts contains per-row options for rendering a pane (future use)
type PaneRowOpts struct {
	Selected bool
}

// Column component functions - each returns a styled string

// SpacerStyle returns a spacer with background when selected
func SpacerStyle(text string, selected bool) string {
	if selected {
		return lipgloss.NewStyle().Background(Colors.Bg.Selected).Render(text)
	}
	return text
}

// RenderIndex renders the session number (1-9)
func RenderIndex(num int, selected bool) string {
	label := fmt.Sprintf("%d", num)
	if selected {
		return IndexSelectedStyle.Render(label)
	}
	return IndexStyle.Render(label)
}

// RenderExpandIcon renders the expand/collapse indicator
func RenderExpandIcon(expanded, selected bool) string {
	if expanded {
		if selected {
			return ExpandedIconSelected
		}
		return ExpandedIcon
	}
	if selected {
		return CollapsedIconSelected
	}
	return CollapsedIcon
}

// RenderName renders a name with padding to a fixed width
func RenderName(name string, width int, selected bool, selectedStyle, defaultStyle lipgloss.Style) string {
	padded := fmt.Sprintf("%-*s", width, name)
	if selected {
		return selectedStyle.Render(padded)
	}
	return defaultStyle.Render(padded)
}

// RenderSessionName renders the session name
func RenderSessionName(name string, width int, selected bool) string {
	return RenderName(name, width, selected, SessionNameSelectedStyle, SessionNameStyle)
}

// RenderWindowName renders the window name with index
func RenderWindowName(index int, name string, selected bool) string {
	text := fmt.Sprintf("%d: %s", index, name)
	if selected {
		return WindowNameSelectedStyle.Render(text)
	}
	return WindowNameStyle.Render(text)
}

// RenderTimeAgo renders the time since last activity
func RenderTimeAgo(t time.Time, selected bool) string {
	timeAgo := FormatTimeAgo(t)
	padded := fmt.Sprintf("%-8s", timeAgo)
	if selected {
		return TimeSelectedStyle.Render(padded)
	}
	return TimeStyle.Render(padded)
}

// FormatTimeAgo formats a time as a human-readable "X ago" string
func FormatTimeAgo(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// RenderGitStatusColumn renders the git status with padding to a fixed width
func RenderGitStatusColumn(status *git.Status, maxWidth int, selected bool) string {
	if maxWidth == 0 {
		return ""
	}

	if status == nil {
		return SpacerStyle(strings.Repeat(" ", maxWidth), selected)
	}

	formatted := FormatGitStatus(status.Dirty, status.Additions, status.Deletions, selected)
	actualWidth := GitStatusWidth(status.Dirty, status.Additions, status.Deletions)

	if actualWidth < maxWidth {
		padding := SpacerStyle(strings.Repeat(" ", maxWidth-actualWidth), selected)
		return formatted + padding
	}
	return formatted
}

// RenderClaudeIcon renders a single-character Claude status icon
// Returns a space for no status to preserve column alignment
func RenderClaudeIcon(status *claude.Status, animFrame int, selected bool) string {
	if status == nil || status.State == "" {
		return SpacerStyle(" ", selected) // Reserved space for alignment
	}
	waitDuration := time.Since(status.Timestamp)
	icon := FormatClaudeIcon(status.State, animFrame, waitDuration)
	if selected {
		// Re-apply the icon style with background
		return lipgloss.NewStyle().Background(Colors.Bg.Selected).Render(icon)
	}
	return icon
}

// SessionRowOpts wraps RowOpts with session-specific settings
type SessionRowOpts struct {
	RowOpts
}

// RenderSessionRow composes all columns into a complete session row
func RenderSessionRow(name string, lastActivity time.Time, layout RowLayout, opts SessionRowOpts, width int) string {
	cols := []string{
		RenderIndex(opts.Num, opts.Selected),
		SpacerStyle(" ", opts.Selected),
		// Claude icon (after index, before expand arrow)
		RenderClaudeIcon(opts.ClaudeStatus, opts.AnimFrame, opts.Selected),
		SpacerStyle(" ", opts.Selected),
	}

	// Expand icon (optional)
	if opts.ShowExpandIcon {
		cols = append(cols, RenderExpandIcon(opts.Expanded, opts.Selected), SpacerStyle(" ", opts.Selected))
	}

	// Name (always shown)
	cols = append(cols, RenderSessionName(name, layout.NameWidth, opts.Selected))

	// Time ago (optional)
	if opts.LastActivity != nil {
		cols = append(cols, SpacerStyle("  ", opts.Selected), RenderTimeAgo(*opts.LastActivity, opts.Selected))
	}

	// Git status (optional column)
	if layout.GitStatusWidth > 0 && opts.GitStatus != nil {
		cols = append(cols, SpacerStyle(" ", opts.Selected), RenderGitStatusColumn(opts.GitStatus, layout.GitStatusWidth, opts.Selected))
	} else if layout.GitStatusWidth > 0 {
		// Pad for alignment even when no git status
		cols = append(cols, SpacerStyle(" ", opts.Selected), SpacerStyle(strings.Repeat(" ", layout.GitStatusWidth), opts.Selected))
	}

	content := strings.Join(cols, "")
	if opts.Selected {
		return SessionSelectedStyle.Width(width).Render(content)
	}
	return SessionStyle.Width(width).Render(content)
}

// RenderBookmarkRow composes a bookmark row (simpler than session row)
func RenderBookmarkRow(name string, layout RowLayout, opts RowOpts, width int) string {
	cols := []string{
		RenderIndex(opts.Num, opts.Selected),
		SpacerStyle(" ", opts.Selected),
		// Claude icon column (always reserved for alignment with session rows)
		RenderClaudeIcon(opts.ClaudeStatus, opts.AnimFrame, opts.Selected),
		SpacerStyle(" ", opts.Selected),
		RenderSessionName(name, layout.NameWidth, opts.Selected),
	}

	// Git status (optional)
	if layout.GitStatusWidth > 0 && opts.GitStatus != nil {
		cols = append(cols, SpacerStyle(" ", opts.Selected), RenderGitStatusColumn(opts.GitStatus, layout.GitStatusWidth, opts.Selected))
	}

	content := strings.Join(cols, "")
	if opts.Selected {
		return SessionSelectedStyle.Width(width).Render(content)
	}
	return SessionStyle.Width(width).Render(content)
}

// TableHeaderOpts controls which columns appear in the header
type TableHeaderOpts struct {
	ShowExpandIcon bool
	ShowTime       bool
	ShowGit        bool
	NameLabel      string // e.g., "Session" or "Bookmark"
}

// RenderTableHeader renders a header row above the content list
func RenderTableHeader(layout RowLayout, opts TableHeaderOpts) string {
	dim := TableHeaderTextStyle // shorthand for dim text

	cols := []string{
		"  ", // Align with scrollbar column in data rows
		dim.Render(fmt.Sprintf("%-3s", "#")),
		" ",
		CCHeaderStyle.Render("CC"), // Claude Code status column (orange)
		" ",
	}

	// Expand icon placeholder
	if opts.ShowExpandIcon {
		cols = append(cols, " ", " ")
	}

	// Name column header
	nameLabel := opts.NameLabel
	if nameLabel == "" {
		nameLabel = "NAME"
	}
	cols = append(cols, dim.Render(fmt.Sprintf("%-*s", layout.NameWidth, nameLabel)))

	// Time column header
	if opts.ShowTime {
		cols = append(cols, "  ", dim.Render(fmt.Sprintf("%-8s", "ACT")))
	}

	// Git column header
	if opts.ShowGit && layout.GitStatusWidth > 0 {
		cols = append(cols, " ", dim.Render(fmt.Sprintf("%-*s", layout.GitStatusWidth, "GIT")))
	}

	content := strings.Join(cols, "")
	return TableHeaderStyle.Render(content)
}

// RenderWindowRow composes a window row
func RenderWindowRow(index int, name string, opts WindowRowOpts, width int) string {
	content := RenderWindowName(index, name, opts.Selected)
	if opts.Selected {
		return WindowSelectedStyle.Width(width).Render(content)
	}
	return WindowStyle.Width(width).Render(content)
}

// RenderPaneRow composes a pane row (future use for helm-xdn)
// Panes will be indented further than windows
func RenderPaneRow(index int, title string, opts PaneRowOpts) string {
	text := fmt.Sprintf("%d: %s", index, title)
	if opts.Selected {
		// TODO: Add PaneSelectedStyle when implementing helm-xdn
		return WindowSelectedStyle.PaddingLeft(14).Render(text)
	}
	// TODO: Add PaneStyle when implementing helm-xdn
	return WindowStyle.PaddingLeft(14).Render(text)
}

// ItemDepth represents the hierarchy level of an item
// This design allows easy extension for panes (helm-xdn)
type ItemDepth int

const (
	DepthSession ItemDepth = 0
	DepthWindow  ItemDepth = 1
	DepthPane    ItemDepth = 2 // Future use
)

// IndentForDepth returns the left padding for a given depth level
func IndentForDepth(depth ItemDepth) int {
	switch depth {
	case DepthSession:
		return 0
	case DepthWindow:
		return 10 // Matches current WindowStyle.PaddingLeft
	case DepthPane:
		return 14 // Further indented for panes
	default:
		return 0
	}
}
