package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/nikbrunner/tsm/internal/claude"
	"github.com/nikbrunner/tsm/internal/git"
)

// RowLayout holds calculated column widths for consistent alignment across rows
type RowLayout struct {
	NameWidth      int
	GitStatusWidth int
}

// SessionRowOpts contains per-row options for rendering a session
type SessionRowOpts struct {
	Num          int
	IsFirst      bool
	Selected     bool
	Expanded     bool
	GitStatus    *git.Status
	ClaudeStatus *claude.Status
	AnimFrame    int
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

// RenderIndex renders the session number (1-9)
func RenderIndex(num int, selected bool) string {
	label := fmt.Sprintf("%d", num)
	if selected {
		return IndexSelectedStyle.Render(label)
	}
	return IndexStyle.Render(label)
}

// RenderLastIcon renders the "last session" indicator
func RenderLastIcon(isFirst, selected bool) string {
	if isFirst {
		if selected {
			return LastIconSelected
		}
		return LastIcon
	}
	return " " // Fixed width placeholder
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
func RenderName(name string, width int, selected bool, style lipgloss.Style) string {
	padded := fmt.Sprintf("%-*s", width, name)
	if selected {
		return style.Render(padded)
	}
	return padded
}

// RenderSessionName renders the session name
func RenderSessionName(name string, width int, selected bool) string {
	return RenderName(name, width, selected, SessionNameSelectedStyle)
}

// RenderWindowName renders the window name with index
func RenderWindowName(index int, name string, selected bool) string {
	text := fmt.Sprintf("%d: %s", index, name)
	if selected {
		return WindowNameSelectedStyle.Render(text)
	}
	return text
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
func RenderGitStatusColumn(status *git.Status, maxWidth int) string {
	if maxWidth == 0 {
		return ""
	}

	if status == nil {
		return strings.Repeat(" ", maxWidth)
	}

	formatted := FormatGitStatus(status.Dirty, status.Ahead, status.Behind)
	actualWidth := GitStatusWidth(status.Dirty, status.Ahead, status.Behind)

	if actualWidth < maxWidth {
		return formatted + strings.Repeat(" ", maxWidth-actualWidth)
	}
	return formatted
}

// RenderClaudeStatusColumn renders the Claude status badge
func RenderClaudeStatusColumn(status *claude.Status, animFrame int) string {
	if status == nil {
		return ""
	}
	return FormatClaudeStatus(status.State, animFrame)
}

// RenderSessionRow composes all columns into a complete session row
func RenderSessionRow(name string, lastActivity time.Time, layout RowLayout, opts SessionRowOpts) string {
	cols := []string{
		RenderIndex(opts.Num, opts.Selected),
		" ",
		RenderLastIcon(opts.IsFirst, opts.Selected),
		" ",
		RenderExpandIcon(opts.Expanded, opts.Selected),
		" ",
		RenderSessionName(name, layout.NameWidth, opts.Selected),
		"  ",
		RenderTimeAgo(lastActivity, opts.Selected),
	}

	// Git status (optional column)
	if layout.GitStatusWidth > 0 {
		cols = append(cols, " ", RenderGitStatusColumn(opts.GitStatus, layout.GitStatusWidth))
	}

	// Claude status (optional column)
	if opts.ClaudeStatus != nil {
		claudeStr := RenderClaudeStatusColumn(opts.ClaudeStatus, opts.AnimFrame)
		if claudeStr != "" {
			cols = append(cols, " ", claudeStr)
		}
	}

	content := strings.Join(cols, "")
	return SessionStyle.Render(content)
}

// RenderWindowRow composes a window row
func RenderWindowRow(index int, name string, opts WindowRowOpts) string {
	content := RenderWindowName(index, name, opts.Selected)
	if opts.Selected {
		return WindowSelectedStyle.Render(content)
	}
	return WindowStyle.Render(content)
}

// RenderPaneRow composes a pane row (future use for tsm-xdn)
// Panes will be indented further than windows
func RenderPaneRow(index int, title string, opts PaneRowOpts) string {
	text := fmt.Sprintf("%d: %s", index, title)
	if opts.Selected {
		// TODO: Add PaneSelectedStyle when implementing tsm-xdn
		return WindowSelectedStyle.PaddingLeft(14).Render(text)
	}
	// TODO: Add PaneStyle when implementing tsm-xdn
	return WindowStyle.PaddingLeft(14).Render(text)
}

// ItemDepth represents the hierarchy level of an item
// This design allows easy extension for panes (tsm-xdn)
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
