package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home := os.Getenv("HOME")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expands tilde",
			input:    "~/foo/bar",
			expected: filepath.Join(home, "foo/bar"),
		},
		{
			name:     "leaves absolute path unchanged",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "leaves relative path unchanged",
			input:    "foo/bar",
			expected: "foo/bar",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles tilde only",
			input:    "~",
			expected: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check that defaults are set
	if cfg.ProjectDepth != 2 {
		t.Errorf("ProjectDepth = %d, want 2", cfg.ProjectDepth)
	}

	if cfg.ClaudeStatusEnabled != false {
		t.Error("ClaudeStatusEnabled should be false by default")
	}

	if cfg.Layout != "" {
		t.Errorf("Layout = %q, want empty string", cfg.Layout)
	}
}

func TestPath(t *testing.T) {
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".config", "tsm", "config.yml")

	result := Path()
	if result != expected {
		t.Errorf("Path() = %q, want %q", result, expected)
	}
}

func TestBookmarksPath(t *testing.T) {
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".config", "tsm", "bookmarks.yml")

	result := BookmarksPath()
	if result != expected {
		t.Errorf("BookmarksPath() = %q, want %q", result, expected)
	}
}

func TestSaveAndLoadBookmarks(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "tsm-config-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Override HOME for this test
	t.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "tsm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a config with bookmarks
	cfg := DefaultConfig()
	cfg.Bookmarks = []Bookmark{
		{Path: "/path/to/project1"},
		{Path: "/path/to/project2"},
	}

	// Save bookmarks
	if err := cfg.SaveBookmarks(); err != nil {
		t.Fatalf("SaveBookmarks() error: %v", err)
	}

	// Verify bookmarks file was created
	bookmarksPath := BookmarksPath()
	if _, err := os.Stat(bookmarksPath); os.IsNotExist(err) {
		t.Fatal("bookmarks.yml was not created")
	}

	// Load bookmarks from the file
	loadedBookmarks, err := LoadBookmarks()
	if err != nil {
		t.Fatalf("LoadBookmarks() error: %v", err)
	}

	if len(loadedBookmarks) != 2 {
		t.Errorf("LoadBookmarks() returned %d bookmarks, want 2", len(loadedBookmarks))
	}

	if loadedBookmarks[0].Path != "/path/to/project1" {
		t.Errorf("First bookmark path = %q, want %q", loadedBookmarks[0].Path, "/path/to/project1")
	}
}

func TestBookmarksFileTakesPriorityOverConfig(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "tsm-config-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Override HOME for this test
	t.Setenv("HOME", tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "tsm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config.yml with one bookmark
	configContent := `layout: test
bookmarks:
  - path: /from/config
`
	if err := os.WriteFile(Path(), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create bookmarks.yml with a different bookmark
	bookmarksContent := `bookmarks:
  - path: /from/bookmarks
`
	if err := os.WriteFile(BookmarksPath(), []byte(bookmarksContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load config - should get bookmarks from bookmarks.yml, not config.yml
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Bookmarks) != 1 {
		t.Fatalf("Expected 1 bookmark, got %d", len(cfg.Bookmarks))
	}

	if cfg.Bookmarks[0].Path != "/from/bookmarks" {
		t.Errorf("Bookmark path = %q, want %q (bookmarks.yml should take priority)", cfg.Bookmarks[0].Path, "/from/bookmarks")
	}
}
