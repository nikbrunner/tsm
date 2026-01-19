package model

import (
	"testing"

	"github.com/black-atom-industries/helm/internal/config"
	"github.com/black-atom-industries/helm/internal/tmux"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		{
			name:    "exact match",
			text:    "hello",
			pattern: "hello",
			want:    true,
		},
		{
			name:    "case insensitive",
			text:    "Hello",
			pattern: "hello",
			want:    true,
		},
		{
			name:    "substring match",
			text:    "hello-world",
			pattern: "world",
			want:    true,
		},
		{
			name:    "no match",
			text:    "hello",
			pattern: "xyz",
			want:    false,
		},
		{
			name:    "empty pattern matches all",
			text:    "hello",
			pattern: "",
			want:    true,
		},
		{
			name:    "empty text with pattern",
			text:    "",
			pattern: "hello",
			want:    false,
		},
		{
			name:    "both empty",
			text:    "",
			pattern: "",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyMatch(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestIsCursorValid(t *testing.T) {
	m := Model{
		items: []Item{
			{Type: ItemTypeSession, SessionIndex: 0},
			{Type: ItemTypeSession, SessionIndex: 1},
			{Type: ItemTypeSession, SessionIndex: 2},
		},
	}

	tests := []struct {
		name   string
		cursor int
		want   bool
	}{
		{name: "valid first", cursor: 0, want: true},
		{name: "valid middle", cursor: 1, want: true},
		{name: "valid last", cursor: 2, want: true},
		{name: "negative", cursor: -1, want: false},
		{name: "out of bounds", cursor: 3, want: false},
		{name: "way out of bounds", cursor: 100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.cursor = tt.cursor
			got := m.isCursorValid()
			if got != tt.want {
				t.Errorf("isCursorValid() with cursor=%d = %v, want %v", tt.cursor, got, tt.want)
			}
		})
	}
}

func TestGetTargetName(t *testing.T) {
	m := Model{
		sessions: []tmux.Session{
			{
				Name: "session1",
				Windows: []tmux.Window{
					{Index: 1, Name: "window1", Panes: []tmux.Pane{{Index: 0, Command: "zsh"}, {Index: 1, Command: "nvim"}}},
					{Index: 2, Name: "window2"},
				},
			},
			{
				Name: "session2",
				Windows: []tmux.Window{
					{Index: 1, Name: "main"},
				},
			},
		},
	}

	tests := []struct {
		name string
		item Item
		want string
	}{
		{
			name: "session item",
			item: Item{Type: ItemTypeSession, SessionIndex: 0},
			want: "session1",
		},
		{
			name: "second session",
			item: Item{Type: ItemTypeSession, SessionIndex: 1},
			want: "session2",
		},
		{
			name: "window item",
			item: Item{Type: ItemTypeWindow, SessionIndex: 0, WindowIndex: 0},
			want: "session1:1",
		},
		{
			name: "second window",
			item: Item{Type: ItemTypeWindow, SessionIndex: 0, WindowIndex: 1},
			want: "session1:2",
		},
		{
			name: "pane item",
			item: Item{Type: ItemTypePane, SessionIndex: 0, WindowIndex: 0, PaneIndex: 0},
			want: "session1:1.0",
		},
		{
			name: "second pane",
			item: Item{Type: ItemTypePane, SessionIndex: 0, WindowIndex: 0, PaneIndex: 1},
			want: "session1:1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.getTargetName(tt.item)
			if got != tt.want {
				t.Errorf("getTargetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetError(t *testing.T) {
	m := Model{}

	m.setError("test error: %d", 42)

	if m.message != "test error: 42" {
		t.Errorf("message = %q, want %q", m.message, "test error: 42")
	}

	if !m.messageIsError {
		t.Error("messageIsError should be true")
	}
}

func TestNew(t *testing.T) {
	cfg := config.DefaultConfig()
	m := New("current-session", cfg)

	if m.currentSession != "current-session" {
		t.Errorf("currentSession = %q, want %q", m.currentSession, "current-session")
	}

	if m.mode != ModeNormal {
		t.Errorf("mode = %v, want ModeNormal", m.mode)
	}
}

func TestSanitizeSessionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name unchanged",
			input: "my-session",
			want:  "my-session",
		},
		{
			name:  "slashes to dashes",
			input: "owner/repo",
			want:  "owner-repo",
		},
		{
			name:  "dots to dashes",
			input: "nbr.haus",
			want:  "nbr-haus",
		},
		{
			name:  "colons to dashes",
			input: "session:window",
			want:  "session-window",
		},
		{
			name:  "mixed special chars",
			input: "owner/repo.name:tag",
			want:  "owner-repo-name-tag",
		},
		{
			name:  "spaces to dashes",
			input: "my session name",
			want:  "my-session-name",
		},
		{
			name:  "real world example",
			input: "nikbrunner/nbr.haus",
			want:  "nikbrunner-nbr-haus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeSessionName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeSessionName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Layout height calculation tests

func TestContentWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
		want  int
	}{
		{
			name:  "zero width returns default",
			width: 0,
			want:  56, // Default fallback (60 - 4)
		},
		{
			name:  "normal width subtracts border overhead",
			width: 80,
			want:  76, // 80 - 4 (AppBorderOverheadX)
		},
		{
			name:  "small width",
			width: 40,
			want:  36, // 40 - 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{width: tt.width}
			got := m.contentWidth()
			if got != tt.want {
				t.Errorf("contentWidth() with width=%d = %d, want %d", tt.width, got, tt.want)
			}
		})
	}
}

func TestContentHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "zero height returns zero",
			height: 0,
			want:   0,
		},
		{
			name:   "normal height subtracts border overhead",
			height: 30,
			want:   28, // 30 - 2 (AppBorderOverheadY)
		},
		{
			name:   "small height",
			height: 10,
			want:   8, // 10 - 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{height: tt.height}
			got := m.contentHeight()
			if got != tt.want {
				t.Errorf("contentHeight() with height=%d = %d, want %d", tt.height, got, tt.want)
			}
		})
	}
}

func TestSessionMaxVisibleItems(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "zero height returns fallback",
			height: 0,
			want:   10,
		},
		{
			name:   "small window",
			height: 12, // contentHeight = 10, available = 10 - 8 = 2
			want:   2,
		},
		{
			name:   "large window uses all space",
			height: 50, // contentHeight = 48, available = 48 - 8 = 40
			want:   40,
		},
		{
			name:   "very small window",
			height: 10, // contentHeight = 8, available = 8 - 8 = 0, returns fallback
			want:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			m := Model{
				height: tt.height,
				config: cfg,
			}
			got := m.sessionMaxVisibleItems()
			if got != tt.want {
				t.Errorf("sessionMaxVisibleItems() with height=%d = %d, want %d",
					tt.height, got, tt.want)
			}
		})
	}
}

func TestProjectMaxVisibleItems(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   int
	}{
		{
			name:   "zero height returns fallback",
			height: 0,
			want:   10,
		},
		{
			name:   "small window",
			height: 12, // contentHeight = 10, available = 10 - 8 = 2
			want:   2,
		},
		{
			name:   "large window uses all space",
			height: 50, // contentHeight = 48, available = 48 - 8 = 40
			want:   40,
		},
		{
			name:   "medium window",
			height: 17, // contentHeight = 15, available = 15 - 8 = 7
			want:   7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			m := Model{
				height: tt.height,
				config: cfg,
			}
			got := m.projectMaxVisibleItems()
			if got != tt.want {
				t.Errorf("projectMaxVisibleItems() with height=%d = %d, want %d",
					tt.height, got, tt.want)
			}
		})
	}
}
