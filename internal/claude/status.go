package claude

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StaleThreshold is how long before a status is considered stale.
// If Claude Code hasn't updated the status file in this time, assume it's not running.
const StaleThreshold = 2 * time.Minute

// Status represents Claude Code status for a session
type Status struct {
	State     string    // "new", "working", "waiting", or ""
	Timestamp time.Time // When the status was last updated
}

// IsStale returns true if the status hasn't been updated within StaleThreshold.
func (s Status) IsStale() bool {
	if s.State == "" {
		return false // No status to be stale
	}
	return time.Since(s.Timestamp) > StaleThreshold
}

// GetStatus reads the Claude Code status for a session from the given cache directory.
// Returns empty Status if no status file exists or if status is stale.
func GetStatus(sessionName string, cacheDir string) Status {
	statusFile := filepath.Join(cacheDir, sessionName+".status")
	content, err := os.ReadFile(statusFile)
	if err != nil {
		return Status{}
	}

	// Parse format: "state:timestamp"
	parts := strings.SplitN(strings.TrimSpace(string(content)), ":", 2)
	if len(parts) != 2 {
		return Status{}
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return Status{}
	}

	status := Status{
		State:     parts[0],
		Timestamp: time.Unix(timestamp, 0),
	}

	// If status is stale, treat it as no status
	if status.IsStale() {
		return Status{}
	}

	return status
}

// CleanupStale removes status files for sessions that no longer exist
func CleanupStale(cacheDir string, activeSessions []string) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}

	activeSet := make(map[string]bool)
	for _, s := range activeSessions {
		activeSet[s] = true
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".status") {
			continue
		}

		sessionName := strings.TrimSuffix(entry.Name(), ".status")
		if !activeSet[sessionName] {
			_ = os.Remove(filepath.Join(cacheDir, entry.Name()))
		}
	}
}
