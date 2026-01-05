package claude

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "claude-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name        string
		filename    string
		content     string
		wantState   string
		wantTimeSet bool
	}{
		{
			name:        "valid working status",
			filename:    "test-session.status",
			content:     "working:1704067200",
			wantState:   "working",
			wantTimeSet: true,
		},
		{
			name:        "valid waiting status",
			filename:    "test-session.status",
			content:     "waiting:1704067200",
			wantState:   "waiting",
			wantTimeSet: true,
		},
		{
			name:        "valid new status",
			filename:    "test-session.status",
			content:     "new:1704067200",
			wantState:   "new",
			wantTimeSet: true,
		},
		{
			name:        "missing file returns empty",
			filename:    "nonexistent.status",
			content:     "",
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "malformed content - no colon",
			filename:    "test-session.status",
			content:     "working",
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "malformed content - invalid timestamp",
			filename:    "test-session.status",
			content:     "working:notanumber",
			wantState:   "",
			wantTimeSet: false,
		},
		{
			name:        "empty file",
			filename:    "test-session.status",
			content:     "",
			wantState:   "",
			wantTimeSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file if content provided
			if tt.content != "" || tt.name == "empty file" {
				filePath := filepath.Join(tmpDir, tt.filename)
				if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				defer func() { _ = os.Remove(filePath) }()
			}

			sessionName := "test-session"
			if tt.name == "missing file returns empty" {
				sessionName = "nonexistent"
			}

			status := GetStatus(sessionName, tmpDir)

			if status.State != tt.wantState {
				t.Errorf("State = %q, want %q", status.State, tt.wantState)
			}

			if tt.wantTimeSet && status.Timestamp.IsZero() {
				t.Error("Timestamp should be set")
			}

			if !tt.wantTimeSet && !status.Timestamp.IsZero() {
				t.Error("Timestamp should be zero")
			}
		})
	}
}

func TestCleanupStale(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "claude-cleanup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create some status files
	files := []string{"active.status", "stale1.status", "stale2.status", "notastatus.txt"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("working:"+string(rune(time.Now().Unix()))), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", f, err)
		}
	}

	// Run cleanup with only "active" as active session
	CleanupStale(tmpDir, []string{"active"})

	// Check that active.status still exists
	if _, err := os.Stat(filepath.Join(tmpDir, "active.status")); os.IsNotExist(err) {
		t.Error("active.status should not be deleted")
	}

	// Check that stale files are deleted
	if _, err := os.Stat(filepath.Join(tmpDir, "stale1.status")); !os.IsNotExist(err) {
		t.Error("stale1.status should be deleted")
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "stale2.status")); !os.IsNotExist(err) {
		t.Error("stale2.status should be deleted")
	}

	// Check that non-status file is not touched
	if _, err := os.Stat(filepath.Join(tmpDir, "notastatus.txt")); os.IsNotExist(err) {
		t.Error("notastatus.txt should not be deleted")
	}
}
