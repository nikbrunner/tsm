package ui

import "strings"

// ScrollList is a generic scrollable list with cursor, filtering, and scroll offset management.
// It eliminates duplicate scroll/cursor logic across different list modes.
type ScrollList[T any] struct {
	items        []T
	filtered     []T
	cursor       int
	scrollOffset int
	filter       string
	filterFn     func(T, string) bool // Returns true if item matches filter
	height       int                  // Visible height (number of items that fit)
}

// NewScrollList creates a new ScrollList with a filter function.
// The filterFn should return true if the item matches the given filter string.
func NewScrollList[T any](filterFn func(T, string) bool) *ScrollList[T] {
	return &ScrollList[T]{
		filterFn: filterFn,
		height:   10, // Default fallback
	}
}

// SetItems replaces all items and re-applies the current filter
func (s *ScrollList[T]) SetItems(items []T) {
	s.items = items
	s.applyFilter()
}

// SetHeight sets the visible height (number of items that fit on screen)
func (s *ScrollList[T]) SetHeight(height int) {
	if height > 0 {
		s.height = height
	}
	s.updateScrollOffset()
}

// Height returns the current visible height
func (s *ScrollList[T]) Height() int {
	return s.height
}

// SetFilter sets the filter text and re-filters the items
func (s *ScrollList[T]) SetFilter(filter string) {
	s.filter = filter
	s.applyFilter()
}

// Filter returns the current filter string
func (s *ScrollList[T]) Filter() string {
	return s.filter
}

// applyFilter applies the current filter to items
func (s *ScrollList[T]) applyFilter() {
	if s.filter == "" {
		s.filtered = s.items
	} else {
		s.filtered = nil
		filterLower := strings.ToLower(s.filter)
		for _, item := range s.items {
			if s.filterFn(item, filterLower) {
				s.filtered = append(s.filtered, item)
			}
		}
	}
	s.clampCursor()
	s.updateScrollOffset()
}

// Items returns all items (unfiltered)
func (s *ScrollList[T]) Items() []T {
	return s.items
}

// Filtered returns the filtered items
func (s *ScrollList[T]) Filtered() []T {
	return s.filtered
}

// Len returns the number of filtered items
func (s *ScrollList[T]) Len() int {
	return len(s.filtered)
}

// Cursor returns the current cursor position
func (s *ScrollList[T]) Cursor() int {
	return s.cursor
}

// SetCursor sets the cursor position and updates scroll offset
func (s *ScrollList[T]) SetCursor(pos int) {
	s.cursor = pos
	s.clampCursor()
	s.updateScrollOffset()
}

// MoveCursor moves the cursor by delta and updates scroll offset
func (s *ScrollList[T]) MoveCursor(delta int) {
	s.cursor += delta
	s.clampCursor()
	s.updateScrollOffset()
}

// clampCursor ensures cursor is within valid bounds
func (s *ScrollList[T]) clampCursor() {
	if s.cursor >= len(s.filtered) {
		s.cursor = len(s.filtered) - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
}

// ScrollOffset returns the current scroll offset
func (s *ScrollList[T]) ScrollOffset() int {
	return s.scrollOffset
}

// updateScrollOffset adjusts scroll offset to keep cursor visible
func (s *ScrollList[T]) updateScrollOffset() {
	// If cursor is above visible area, scroll up
	if s.cursor < s.scrollOffset {
		s.scrollOffset = s.cursor
	}
	// If cursor is below visible area, scroll down
	if s.cursor >= s.scrollOffset+s.height {
		s.scrollOffset = s.cursor - s.height + 1
	}
	// Ensure scroll offset is not negative
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
}

// SelectedItem returns the currently selected item, or false if none
func (s *ScrollList[T]) SelectedItem() (T, bool) {
	var zero T
	if s.cursor < 0 || s.cursor >= len(s.filtered) {
		return zero, false
	}
	return s.filtered[s.cursor], true
}

// VisibleItems returns the slice of items currently visible on screen
func (s *ScrollList[T]) VisibleItems() []T {
	if len(s.filtered) == 0 {
		return nil
	}

	start := s.scrollOffset
	end := start + s.height
	if end > len(s.filtered) {
		end = len(s.filtered)
	}
	if start >= end {
		return nil
	}
	return s.filtered[start:end]
}

// VisibleRange returns the start and end indices of visible items in the filtered slice
func (s *ScrollList[T]) VisibleRange() (start, end int) {
	start = s.scrollOffset
	end = start + s.height
	if end > len(s.filtered) {
		end = len(s.filtered)
	}
	return start, end
}

// IsSelected returns true if the given index (in filtered list) is the cursor position
func (s *ScrollList[T]) IsSelected(index int) bool {
	return index == s.cursor
}

// Reset clears the filter and resets cursor to 0
func (s *ScrollList[T]) Reset() {
	s.filter = ""
	s.cursor = 0
	s.scrollOffset = 0
	s.applyFilter()
}

// Clear removes all items
func (s *ScrollList[T]) Clear() {
	s.items = nil
	s.filtered = nil
	s.cursor = 0
	s.scrollOffset = 0
	s.filter = ""
}
