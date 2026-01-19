package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Session represents a tmux session
type Session struct {
	Name         string
	LastActivity time.Time
	Windows      []Window
	Expanded     bool
}

// Window represents a tmux window
type Window struct {
	Index    int
	Name     string
	Panes    []Pane
	Expanded bool
}

// Pane represents a tmux pane
type Pane struct {
	Index   int
	Command string // Current command running in the pane
	Active  bool   // Active pane in the window
}

// CurrentSession returns the name of the current tmux session
func CurrentSession() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ListSessions returns all tmux sessions sorted by activity (most recent first)
// Excludes the current session and popup sessions
func ListSessions(excludeCurrent string) ([]Session, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_activity} #{session_name}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Session{}, nil
	}

	var sessions []Session

	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[1]

		// Skip current session and popup sessions
		if name == excludeCurrent || strings.HasPrefix(name, "_popup_") {
			continue
		}

		activityUnix, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		sessions = append(sessions, Session{
			Name:         name,
			LastActivity: time.Unix(activityUnix, 0),
		})
	}

	// Sort by activity (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}

// ListWindows returns all windows for a given session
func ListWindows(sessionName string) ([]Window, error) {
	out, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}:#{window_name}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Window{}, nil
	}

	var windows []Window
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		index, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		windows = append(windows, Window{
			Index: index,
			Name:  parts[1],
		})
	}

	return windows, nil
}

// KillSession kills a tmux session by name
func KillSession(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

// KillWindow kills a tmux window
func KillWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	return exec.Command("tmux", "kill-window", "-t", target).Run()
}

// SessionExists checks if a tmux session with the given name exists
func SessionExists(name string) bool {
	return exec.Command("tmux", "has-session", "-t", name).Run() == nil
}

// CreateSession creates a new tmux session
func CreateSession(name, dir string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name, "-c", dir).Run()
}

// SwitchClient switches the tmux client to a session or window.
// If running inside tmux, uses switch-client. If outside, uses attach-session.
func SwitchClient(target string) error {
	var cmd *exec.Cmd
	if os.Getenv("TMUX") != "" {
		cmd = exec.Command("tmux", "switch-client", "-t", target)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", target)
		// Connect terminal for interactive attach
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// SelectWindow selects a specific window in the current client
func SelectWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}

// ListPanes returns all panes for a given session and window
func ListPanes(sessionName string, windowIndex int) ([]Pane, error) {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	out, err := exec.Command("tmux", "list-panes", "-t", target, "-F", "#{pane_index}:#{pane_current_command}:#{pane_active}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Pane{}, nil
	}

	var panes []Pane
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		index, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		active := parts[2] == "1"

		panes = append(panes, Pane{
			Index:   index,
			Command: parts[1],
			Active:  active,
		})
	}

	return panes, nil
}

// KillPane kills a tmux pane
func KillPane(sessionName string, windowIndex, paneIndex int) error {
	target := fmt.Sprintf("%s:%d.%d", sessionName, windowIndex, paneIndex)
	return exec.Command("tmux", "kill-pane", "-t", target).Run()
}

// SelectPane switches to a specific pane
func SelectPane(sessionName string, windowIndex, paneIndex int) error {
	target := fmt.Sprintf("%s:%d.%d", sessionName, windowIndex, paneIndex)
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}
