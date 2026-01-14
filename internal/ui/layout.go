package ui

// Layout height constants for UI sections.
// These constants define the fixed overhead for each part of the UI,
// making it easy to calculate available space for content.
const (
	// Header section (above content)
	TitleBarHeight  = 1
	PromptHeight    = 1
	TopBorderHeight = 1
	HeaderOverhead  = 3 // TitleBarHeight + PromptHeight + TopBorderHeight

	// Footer section (below content)
	BottomBorderHeight = 1
	NotificationHeight = 1
	StateLineHeight    = 1
	HintsHeight        = 2
	FooterOverhead     = 5 // BottomBorderHeight + NotificationHeight + StateLineHeight + HintsHeight

	// Optional content elements
	TableHeaderHeight = 1

	// Computed totals for visible items calculation
	BaseOverhead            = 8 // HeaderOverhead + FooterOverhead
	WithTableHeaderOverhead = 9 // BaseOverhead + TableHeaderHeight

	// Fallback values
	DefaultVisibleItems = 10
)
