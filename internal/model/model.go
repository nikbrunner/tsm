package model

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/tsm/internal/claude"
	"github.com/nikbrunner/tsm/internal/config"
	"github.com/nikbrunner/tsm/internal/git"
	"github.com/nikbrunner/tsm/internal/github"
	"github.com/nikbrunner/tsm/internal/repos"
	"github.com/nikbrunner/tsm/internal/tmux"
	"github.com/nikbrunner/tsm/internal/ui"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirmKill
	ModeCreate
	ModePickDirectory
	ModeConfirmRemoveFolder
	ModeCloneRepo
)

// Item represents either a session or a window in the flattened list
type Item struct {
	IsSession    bool
	SessionIndex int // Index in the sessions slice
	WindowIndex  int // Index in the session's windows slice (only for windows)
}

// Model is the main application state
type Model struct {
	sessions          []tmux.Session
	claudeStatuses    map[string]claude.Status
	gitStatuses       map[string]git.Status
	currentSession    string
	cursor            int
	items             []Item // Flattened list of visible items
	mode              Mode
	message           string
	messageIsError    bool
	input             textinput.Model
	killTarget        string // Name of session/window being killed
	removeTarget      string // Full path of folder being removed
	config            config.Config
	maxNameWidth      int    // For column alignment
	maxGitStatusWidth int    // For git status column alignment
	filter            string // Current filter text for fuzzy matching

	// Directory picker state
	projectDirs     []string // All scanned directories
	projectFiltered []string // Filtered list based on projectFilter
	projectFilter   string   // Current filter text for directory picker
	projectCursor   int      // Selected item in directory list

	// Scroll state
	scrollOffset        int // Scroll offset for session list
	projectScrollOffset int // Scroll offset for directory picker

	// Window size
	width  int
	height int

	// Animation state
	animationFrame int

	// Clone repo mode state
	cloneRepos          []string // Available repos to clone (filtered)
	cloneReposAll       []string // All available repos (unfiltered)
	cloneFilter         string   // Current filter text
	cloneCursor         int      // Selected repo index
	cloneScrollOffset   int      // Scroll offset
	cloneBasePath       string   // From repos config
	cloneLoading        bool     // True while fetching repos
	cloneError          string   // Error message if fetch/clone fails
	cloneCloning        bool     // True while cloning
	cloneCloningRepo    string   // Repo being cloned
	cloneSuccess        bool     // True when clone completed, awaiting confirmation
	cloneSuccessPath    string   // Path of cloned repo (for layout)
	cloneSuccessSession string   // Session name to switch to
}

// New creates a new Model
func New(currentSession string, cfg config.Config) Model {
	ti := textinput.New()
	ti.CharLimit = 50

	return Model{
		currentSession: currentSession,
		input:          ti,
		config:         cfg,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadSessions, animationTick())
}

// loadSessions fetches sessions from tmux
func (m Model) loadSessions() tea.Msg {
	sessions, err := tmux.ListSessions(m.currentSession)
	if err != nil {
		return errMsg{err}
	}
	return sessionsMsg{sessions}
}

type sessionsMsg struct {
	sessions []tmux.Session
}

type errMsg struct {
	err error
}

type clearMessageMsg struct{}

type animationTickMsg struct{}

// Clone repo mode messages
type cloneReposLoadedMsg struct {
	repos []string
}

type cloneErrorMsg struct {
	err error
}

type cloneSuccessMsg struct {
	repoPath    string
	sessionName string
}

// clearMessageAfter returns a command that clears the message after a delay
func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

// animationTick returns a command that ticks the animation
func animationTick() tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsMsg:
		m.sessions = msg.sessions
		m.loadClaudeStatuses()
		m.loadGitStatuses()
		m.calculateColumnWidths()
		m.rebuildItems()
		if len(m.items) == 0 {
			m.message = "No other sessions. Press c to create one."
		}
		return m, nil

	case errMsg:
		m.setError("Error: %v", msg.err)
		return m, nil

	case clearMessageMsg:
		m.message = ""
		m.messageIsError = false
		return m, nil

	case animationTickMsg:
		m.animationFrame = (m.animationFrame + 1) % 3
		return m, animationTick()

	case cloneReposLoadedMsg:
		m.cloneLoading = false
		m.cloneReposAll = msg.repos
		m.cloneRepos = msg.repos
		m.cloneCursor = 0
		m.cloneScrollOffset = 0
		if len(msg.repos) == 0 {
			m.cloneError = "All repositories are already cloned!"
		}
		return m, nil

	case cloneErrorMsg:
		m.cloneLoading = false
		m.cloneCloning = false
		m.cloneError = msg.err.Error()
		return m, nil

	case cloneSuccessMsg:
		// Store success state and await confirmation
		m.cloneCloning = false
		m.cloneSuccess = true
		m.cloneSuccessPath = msg.repoPath
		m.cloneSuccessSession = msg.sessionName
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Handle text input updates in create mode
	if m.mode == ModeCreate {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeConfirmKill:
		return m.handleConfirmKillMode(msg)
	case ModeCreate:
		return m.handleCreateMode(msg)
	case ModePickDirectory:
		return m.handlePickDirectoryMode(msg)
	case ModeConfirmRemoveFolder:
		return m.handleConfirmRemoveFolderMode(msg)
	case ModeCloneRepo:
		return m.handleCloneRepoMode(msg)
	}
	return m, nil
}

func (m *Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Cancel):
		// Escape: clear filter if active, otherwise quit
		if m.filter != "" {
			m.filter = ""
			m.rebuildItems()
			return m, nil
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
			m.updateScrollOffset()
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
			m.updateScrollOffset()
		}

	case key.Matches(msg, keys.Expand):
		m.expandCurrent()

	case key.Matches(msg, keys.Collapse):
		m.collapseCurrent()

	case key.Matches(msg, keys.Select):
		return m.selectCurrent()

	case key.Matches(msg, keys.Kill):
		return m.confirmKill()

	case key.Matches(msg, keys.Create):
		m.mode = ModeCreate
		m.filter = "" // Clear any active filter
		// Reset input completely
		m.input.Reset()
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.PickDirectory):
		m.mode = ModePickDirectory
		m.filter = "" // Clear any active filter
		m.projectFilter = ""
		m.projectCursor = 0
		m.projectScrollOffset = 0
		m.projectDirs = m.scanProjectDirectories()
		m.projectFiltered = m.projectDirs
		// Request window size to get proper height for layout
		return m, tea.WindowSize()

	case key.Matches(msg, keys.CloneRepo):
		// Load repos config
		cfg, err := repos.LoadConfig()
		if err != nil {
			m.setError("Failed to load repos config: %v", err)
			return m, nil
		}
		m.cloneBasePath = cfg.ReposBasePath
		m.mode = ModeCloneRepo
		m.filter = "" // Clear any active filter
		m.cloneFilter = ""
		m.cloneCursor = 0
		m.cloneScrollOffset = 0
		m.cloneRepos = nil
		m.cloneReposAll = nil
		m.cloneError = ""
		m.cloneLoading = true
		m.cloneCloning = false
		return m, m.fetchAvailableReposCmd()

	case key.Matches(msg, keys.Lazygit):
		return m.openLazygit()

	// Number jumps (only when no filter active)
	case m.filter == "" && key.Matches(msg, keys.Jump1):
		return m.handleJump(1)
	case m.filter == "" && key.Matches(msg, keys.Jump2):
		return m.handleJump(2)
	case m.filter == "" && key.Matches(msg, keys.Jump3):
		return m.handleJump(3)
	case m.filter == "" && key.Matches(msg, keys.Jump4):
		return m.handleJump(4)
	case m.filter == "" && key.Matches(msg, keys.Jump5):
		return m.handleJump(5)
	case m.filter == "" && key.Matches(msg, keys.Jump6):
		return m.handleJump(6)
	case m.filter == "" && key.Matches(msg, keys.Jump7):
		return m.handleJump(7)
	case m.filter == "" && key.Matches(msg, keys.Jump8):
		return m.handleJump(8)
	case m.filter == "" && key.Matches(msg, keys.Jump9):
		return m.handleJump(9)

	case msg.Type == tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildItems()
		}

	case msg.Type == tea.KeyRunes:
		// Add typed characters to filter
		m.filter += string(msg.Runes)
		m.rebuildItems()
	}

	return m, nil
}

func (m *Model) handleConfirmKillMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Kill):
		// Double C-x confirms the kill
		return m.killCurrent()
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.message = ""
		m.killTarget = ""
	}

	return m, nil
}

func (m *Model) handleCreateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil

	case msg.Type == tea.KeyEnter:
		name := strings.TrimSpace(m.input.Value())
		if name == "" {
			m.setError("Session name cannot be empty")
			return m, nil
		}
		return m.createSession(name)
	}

	// Ignore ctrl key combinations - only pass regular typing to input
	if msg.Type == tea.KeyCtrlN || msg.Type == tea.KeyCtrlO ||
		msg.Type == tea.KeyCtrlJ || msg.Type == tea.KeyCtrlK ||
		msg.Type == tea.KeyCtrlH || msg.Type == tea.KeyCtrlL ||
		msg.Type == tea.KeyCtrlX || msg.Type == tea.KeyCtrlY ||
		msg.Type == tea.KeyCtrlP {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handlePickDirectoryMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		// Clear filter first, then exit on second press
		if m.projectFilter != "" {
			m.projectFilter = ""
			m.projectFiltered = m.projectDirs
			m.projectCursor = 0
			return m, nil
		}
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.projectCursor > 0 {
			m.projectCursor--
			m.updateProjectScrollOffset()
		}

	case key.Matches(msg, keys.Down):
		if m.projectCursor < len(m.projectFiltered)-1 {
			m.projectCursor++
			m.updateProjectScrollOffset()
		}

	case key.Matches(msg, keys.Select):
		if len(m.projectFiltered) > 0 && m.projectCursor < len(m.projectFiltered) {
			return m.createSessionFromDir(m.projectFiltered[m.projectCursor])
		}

	case key.Matches(msg, keys.Kill):
		return m.confirmRemoveFolder()

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyBackspace:
		if len(m.projectFilter) > 0 {
			m.projectFilter = m.projectFilter[:len(m.projectFilter)-1]
			m.filterProjectDirs()
		}

	case msg.Type == tea.KeyRunes:
		m.projectFilter += string(msg.Runes)
		m.filterProjectDirs()
	}

	return m, nil
}

func (m *Model) handleCloneRepoMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	// Handle confirmation state
	if m.cloneSuccess {
		switch {
		case key.Matches(msg, keys.Select):
			// Apply layout and switch to the session
			m.applyLayout(m.cloneSuccessSession, m.cloneSuccessPath)
			if err := tmux.SwitchClient(m.cloneSuccessSession); err != nil {
				m.setError("Created but failed to switch: %v", err)
				m.mode = ModeNormal
				m.cloneSuccess = false
				return m, m.loadSessions
			}
			return m, tea.Quit

		case key.Matches(msg, keys.Cancel):
			// Go back to session list without switching
			m.mode = ModeNormal
			m.cloneSuccess = false
			return m, m.loadSessions
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, keys.Cancel):
		// If loading or cloning, just cancel and go back
		if m.cloneLoading || m.cloneCloning {
			m.mode = ModeNormal
			m.cloneLoading = false
			m.cloneCloning = false
			m.cloneError = ""
			return m, nil
		}
		// Clear filter first, then exit on second press
		if m.cloneFilter != "" {
			m.cloneFilter = ""
			m.cloneRepos = m.cloneReposAll
			m.cloneCursor = 0
			m.cloneScrollOffset = 0
			return m, nil
		}
		// If there's an error, clear it and go back
		if m.cloneError != "" {
			m.mode = ModeNormal
			m.cloneError = ""
			return m, nil
		}
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.cloneCursor > 0 {
			m.cloneCursor--
			m.updateCloneScrollOffset()
		}

	case key.Matches(msg, keys.Down):
		if m.cloneCursor < len(m.cloneRepos)-1 {
			m.cloneCursor++
			m.updateCloneScrollOffset()
		}

	case key.Matches(msg, keys.Select):
		if len(m.cloneRepos) > 0 && m.cloneCursor < len(m.cloneRepos) && !m.cloneLoading && !m.cloneCloning && m.cloneError == "" {
			return m.cloneSelectedRepo()
		}

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyBackspace:
		if len(m.cloneFilter) > 0 && !m.cloneLoading && !m.cloneCloning {
			m.cloneFilter = m.cloneFilter[:len(m.cloneFilter)-1]
			m.filterCloneRepos()
		}

	case msg.Type == tea.KeyRunes:
		if !m.cloneLoading && !m.cloneCloning && m.cloneError == "" {
			m.cloneFilter += string(msg.Runes)
			m.filterCloneRepos()
		}
	}

	return m, nil
}

// filterCloneRepos filters the clone repos based on cloneFilter
func (m *Model) filterCloneRepos() {
	if m.cloneFilter == "" {
		m.cloneRepos = m.cloneReposAll
	} else {
		filterLower := strings.ToLower(m.cloneFilter)
		m.cloneRepos = nil
		for _, repo := range m.cloneReposAll {
			if fuzzyMatch(repo, filterLower) {
				m.cloneRepos = append(m.cloneRepos, repo)
			}
		}
	}
	// Reset cursor if out of bounds
	if m.cloneCursor >= len(m.cloneRepos) {
		m.cloneCursor = len(m.cloneRepos) - 1
	}
	if m.cloneCursor < 0 {
		m.cloneCursor = 0
	}
	m.updateCloneScrollOffset()
}

// updateCloneScrollOffset adjusts scroll offset to keep cursor visible
func (m *Model) updateCloneScrollOffset() {
	maxVisible := m.cloneMaxVisibleItems()
	if m.cloneCursor < m.cloneScrollOffset {
		m.cloneScrollOffset = m.cloneCursor
	}
	if m.cloneCursor >= m.cloneScrollOffset+maxVisible {
		m.cloneScrollOffset = m.cloneCursor - maxVisible + 1
	}
	if m.cloneScrollOffset < 0 {
		m.cloneScrollOffset = 0
	}
}

// cloneMaxVisibleItems returns the number of items visible in clone mode
func (m *Model) cloneMaxVisibleItems() int {
	contentH := m.contentHeight()
	if contentH > 0 {
		availableForContent := contentH - 5
		if availableForContent > 0 {
			return availableForContent
		}
	}
	// Fallback when height unknown
	return 10
}

// cloneSelectedRepo starts cloning the selected repository
func (m *Model) cloneSelectedRepo() (tea.Model, tea.Cmd) {
	selected := m.cloneRepos[m.cloneCursor]
	m.cloneCloning = true
	m.cloneCloningRepo = selected

	destPath := filepath.Join(m.cloneBasePath, selected)
	sessionName := sanitizeSessionName(selected)

	return m, func() tea.Msg {
		if err := github.CloneRepo(selected, destPath); err != nil {
			return cloneErrorMsg{err: err}
		}

		// Create tmux session
		if err := tmux.CreateSession(sessionName, destPath); err != nil {
			return cloneErrorMsg{err: fmt.Errorf("cloned but failed to create session: %w", err)}
		}

		return cloneSuccessMsg{
			repoPath:    destPath,
			sessionName: sessionName,
		}
	}
}

// fetchAvailableReposCmd fetches repos from GitHub
func (m *Model) fetchAvailableReposCmd() tea.Cmd {
	basePath := m.cloneBasePath
	return func() tea.Msg {
		// Check gh CLI
		if err := github.CheckGhCli(); err != nil {
			return cloneErrorMsg{err: err}
		}

		// Fetch available repos
		available, err := github.FetchAvailableRepos()
		if err != nil {
			return cloneErrorMsg{err: err}
		}

		// Get already cloned
		cloned, _ := repos.ListClonedRepos(basePath)

		// Filter out cloned
		uncloned := repos.FilterUncloned(available, cloned)

		return cloneReposLoadedMsg{repos: uncloned}
	}
}

// filterProjectDirs filters the project directories based on projectFilter
// Filters match against the display path (last N components), not full path
func (m *Model) filterProjectDirs() {
	if m.projectFilter == "" {
		m.projectFiltered = m.projectDirs
	} else {
		filterLower := strings.ToLower(m.projectFilter)
		m.projectFiltered = nil
		for _, fullPath := range m.projectDirs {
			displayPath := m.extractDisplayPath(fullPath)
			if fuzzyMatch(displayPath, filterLower) {
				m.projectFiltered = append(m.projectFiltered, fullPath)
			}
		}
	}
	// Reset cursor if out of bounds
	if m.projectCursor >= len(m.projectFiltered) {
		m.projectCursor = len(m.projectFiltered) - 1
	}
	if m.projectCursor < 0 {
		m.projectCursor = 0
	}
	m.updateProjectScrollOffset()
}

func (m *Model) createSessionFromDir(fullPath string) (tea.Model, tea.Cmd) {
	// Extract session name from full path (last N components based on depth)
	name := m.extractSessionName(fullPath)

	// Check if session already exists - if so, just switch to it
	if tmux.SessionExists(name) {
		if err := tmux.SwitchClient(name); err != nil {
			m.setError("Failed to switch: %v", err)
			return m, m.loadSessions
		}
		return m, tea.Quit
	}

	if err := tmux.CreateSession(name, fullPath); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, fullPath)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

// extractSessionName extracts a session name from a full path
// Uses the last N path components based on ProjectDepth config
func (m *Model) extractSessionName(fullPath string) string {
	parts := strings.Split(fullPath, string(filepath.Separator))
	depth := m.config.ProjectDepth
	if depth > len(parts) {
		depth = len(parts)
	}
	relPath := strings.Join(parts[len(parts)-depth:], "/")
	return sanitizeSessionName(relPath)
}

// extractDisplayPath extracts a display path from a full path
// Uses the last N path components based on ProjectDepth config
func (m *Model) extractDisplayPath(fullPath string) string {
	parts := strings.Split(fullPath, string(filepath.Separator))
	depth := m.config.ProjectDepth
	if depth > len(parts) {
		depth = len(parts)
	}
	return strings.Join(parts[len(parts)-depth:], "/")
}

// scanProjectDirectories scans all configured project directories at the configured depth
// and returns full paths to each discovered directory
func (m *Model) scanProjectDirectories() []string {
	var dirs []string
	depth := m.config.ProjectDepth

	// Scan each configured base directory
	for _, baseDir := range m.config.ProjectDirs {
		m.walkAtDepth(baseDir, "", depth, &dirs)
	}

	return dirs
}

// walkAtDepth recursively walks directories and collects full paths at the target depth
func (m *Model) walkAtDepth(baseDir, currentPath string, remainingDepth int, dirs *[]string) {
	if remainingDepth == 0 {
		// We've reached the target depth - add the full path
		if currentPath != "" {
			fullPath := filepath.Join(baseDir, currentPath)
			*dirs = append(*dirs, fullPath)
		}
		return
	}

	// Read the current directory
	scanPath := filepath.Join(baseDir, currentPath)
	entries, err := os.ReadDir(scanPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		var nextPath string
		if currentPath == "" {
			nextPath = entry.Name()
		} else {
			nextPath = filepath.Join(currentPath, entry.Name())
		}

		m.walkAtDepth(baseDir, nextPath, remainingDepth-1, dirs)
	}
}

func (m *Model) handleJump(num int) (tea.Model, tea.Cmd) {
	// Check if we're inside an expanded session - numbers switch to windows
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := m.items[m.cursor]
		session := &m.sessions[item.SessionIndex]

		if session.Expanded {
			// Jump to window number within this session
			for _, w := range session.Windows {
				if w.Index == num {
					target := fmt.Sprintf("%s:%d", session.Name, w.Index)
					if err := tmux.SwitchClient(target); err != nil {
						m.setError("Error: %v", err)
						return m, nil
					}
					return m, tea.Quit
				}
			}
		}
	}

	// Session labels: 1, 2, 3... map to session indices 0, 1, 2...
	sessionIdx := num - 1
	if sessionIdx >= 0 && sessionIdx < len(m.sessions) {
		session := m.sessions[sessionIdx]
		if err := tmux.SwitchClient(session.Name); err != nil {
			m.setError("Error: %v", err)
			return m, nil
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) expandCurrent() {
	if !m.isCursorValid() {
		return
	}

	item := m.items[m.cursor]
	if !item.IsSession {
		return
	}

	// Collapse all other sessions first
	for i := range m.sessions {
		m.sessions[i].Expanded = false
	}

	session := &m.sessions[item.SessionIndex]
	if len(session.Windows) == 0 {
		// Load windows
		windows, err := tmux.ListWindows(session.Name)
		if err != nil {
			m.setError("Error loading windows: %v", err)
			return
		}
		session.Windows = windows
	}
	session.Expanded = true
	m.rebuildItems()
}

func (m *Model) collapseCurrent() {
	if !m.isCursorValid() {
		return
	}

	item := m.items[m.cursor]

	var sessionIdx int
	if item.IsSession {
		sessionIdx = item.SessionIndex
	} else {
		// Collapse parent session
		sessionIdx = item.SessionIndex
		// Move cursor to the session
		for i, it := range m.items {
			if it.IsSession && it.SessionIndex == sessionIdx {
				m.cursor = i
				break
			}
		}
	}

	m.sessions[sessionIdx].Expanded = false
	m.rebuildItems()
}

func (m *Model) selectCurrent() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	target := m.getTargetName(m.items[m.cursor])
	if err := tmux.SwitchClient(target); err != nil {
		m.setError("Error: %v", err)
		return m, nil
	}

	return m, tea.Quit
}

func (m *Model) openLazygit() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	if !item.IsSession {
		// For windows, use the parent session
		item = Item{IsSession: true, SessionIndex: item.SessionIndex}
	}

	session := m.sessions[item.SessionIndex]
	path, err := git.GetSessionPath(session.Name)
	if err != nil || path == "" {
		m.setError("Could not get session path")
		return m, nil
	}

	// Schedule lazygit popup to open after tsm closes, then reopen tsm with same dimensions
	cmd := fmt.Sprintf("sleep 0.1 && tmux display-popup -w%s -h%s -d '%s' -E lazygit; tmux display-popup -w%d -h%d -B -E tsm",
		m.config.LazygitPopup.Width, m.config.LazygitPopup.Height, path, m.width, m.height)
	_ = exec.Command("tmux", "run-shell", "-b", cmd).Start()

	return m, tea.Quit
}

func (m *Model) confirmKill() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	m.killTarget = m.getTargetName(item)

	if item.IsSession {
		m.message = fmt.Sprintf("Kill \"%s\"?", m.killTarget)
	} else {
		m.message = fmt.Sprintf("Kill window \"%s\"?", m.killTarget)
	}

	m.mode = ModeConfirmKill
	return m, nil
}

func (m *Model) killCurrent() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	var err error

	if item.IsSession {
		session := m.sessions[item.SessionIndex]
		err = tmux.KillSession(session.Name)
		if err == nil {
			m.message = fmt.Sprintf("Killed \"%s\"", session.Name)
		}
	} else {
		session := m.sessions[item.SessionIndex]
		window := session.Windows[item.WindowIndex]
		err = tmux.KillWindow(session.Name, window.Index)
		if err == nil {
			m.message = fmt.Sprintf("Killed window %d", window.Index)
		}
	}

	if err != nil {
		m.setError("Error: %v", err)
	}

	m.mode = ModeNormal
	m.killTarget = ""

	// Reload sessions and clear message after 5 seconds
	return m, tea.Batch(m.loadSessions, clearMessageAfter(5*time.Second))
}

func (m *Model) handleConfirmRemoveFolderMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Kill):
		return m.removeFolder()
	case key.Matches(msg, keys.Cancel):
		m.mode = ModePickDirectory
		m.message = ""
		m.removeTarget = ""
	}

	return m, nil
}

func (m *Model) confirmRemoveFolder() (tea.Model, tea.Cmd) {
	if len(m.projectFiltered) == 0 || m.projectCursor >= len(m.projectFiltered) {
		return m, nil
	}

	m.removeTarget = m.projectFiltered[m.projectCursor]
	displayPath := m.extractDisplayPath(m.removeTarget)
	m.message = fmt.Sprintf("Remove \"%s\" from disk?", displayPath)
	m.mode = ModeConfirmRemoveFolder
	return m, nil
}

func (m *Model) removeFolder() (tea.Model, tea.Cmd) {
	if m.removeTarget == "" {
		return m, nil
	}

	displayPath := m.extractDisplayPath(m.removeTarget)
	sessionName := m.extractSessionName(m.removeTarget)

	// Kill associated session if it exists
	if tmux.SessionExists(sessionName) {
		_ = tmux.KillSession(sessionName)
	}

	if err := os.RemoveAll(m.removeTarget); err != nil {
		m.setError("Failed to remove: %v", err)
		m.mode = ModePickDirectory
		m.removeTarget = ""
		return m, nil
	}

	m.message = fmt.Sprintf("Removed \"%s\"", displayPath)
	m.mode = ModePickDirectory
	m.removeTarget = ""

	// Rescan directories and clear message after delay
	m.projectDirs = m.scanProjectDirectories()
	m.filterProjectDirs()

	return m, clearMessageAfter(5 * time.Second)
}

func (m *Model) createSession(name string) (tea.Model, tea.Cmd) {
	// Sanitize session name (spaces, dots, colons break tmux target syntax)
	name = sanitizeSessionName(name)
	workingDir := m.config.DefaultSessionDir
	if err := tmux.CreateSession(name, workingDir); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, workingDir)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

func (m *Model) applyLayout(sessionName, workingDir string) {
	if m.config.Layout == "" {
		return
	}

	scriptPath := fmt.Sprintf("%s/%s.sh", m.config.LayoutDir, m.config.Layout)
	if _, err := os.Stat(scriptPath); err != nil {
		return
	}

	// Run layout script synchronously before switching to the session
	cmd := exec.Command(scriptPath, sessionName, workingDir)
	cmd.Env = append(os.Environ(),
		"TMUX_SESSION="+sessionName,
		"TMUX_WORKING_DIR="+workingDir,
	)
	_ = cmd.Run()
}

func (m *Model) loadClaudeStatuses() {
	m.claudeStatuses = make(map[string]claude.Status)
	if !m.config.ClaudeStatusEnabled {
		return
	}
	for _, s := range m.sessions {
		status := claude.GetStatus(s.Name, m.config.CacheDir)
		if status.State != "" {
			m.claudeStatuses[s.Name] = status
		}
	}
}

func (m *Model) loadGitStatuses() {
	m.gitStatuses = make(map[string]git.Status)
	m.maxGitStatusWidth = 0
	if !m.config.GitStatusEnabled {
		return
	}
	for _, s := range m.sessions {
		path, err := git.GetSessionPath(s.Name)
		if err != nil || path == "" {
			continue
		}
		status := git.GetStatus(path)
		if status.IsRepo && !status.IsClean() {
			m.gitStatuses[s.Name] = status
		}
	}
	// Use fixed column width if any git statuses exist
	if len(m.gitStatuses) > 0 {
		m.maxGitStatusWidth = ui.GitStatusColumnWidth
	}
}

func (m *Model) calculateColumnWidths() {
	m.maxNameWidth = 0
	for _, s := range m.sessions {
		if len(s.Name) > m.maxNameWidth {
			m.maxNameWidth = len(s.Name)
		}
	}
}

func (m *Model) rebuildItems() {
	m.items = nil
	filterLower := strings.ToLower(m.filter)

	for i, session := range m.sessions {
		// Apply fuzzy filter if active
		if m.filter != "" && !fuzzyMatch(session.Name, filterLower) {
			continue
		}

		m.items = append(m.items, Item{
			IsSession:    true,
			SessionIndex: i,
		})

		if session.Expanded {
			for j := range session.Windows {
				m.items = append(m.items, Item{
					IsSession:    false,
					SessionIndex: i,
					WindowIndex:  j,
				})
			}
		}
	}

	// Ensure cursor is in bounds
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.updateScrollOffset()
}

// updateScrollOffset adjusts scroll offset to keep cursor visible in session list
func (m *Model) updateScrollOffset() {
	maxVisible := m.sessionMaxVisibleItems()
	// If cursor is above visible area, scroll up
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	// If cursor is below visible area, scroll down
	if m.cursor >= m.scrollOffset+maxVisible {
		m.scrollOffset = m.cursor - maxVisible + 1
	}
	// Ensure scroll offset is not negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// contentWidth returns the available width inside the app border/padding
func (m *Model) contentWidth() int {
	if m.width > 0 {
		return m.width - ui.AppBorderOverheadX
	}
	return 56 // Default fallback (60 - 4)
}

// contentHeight returns the available height inside the app border/padding
func (m *Model) contentHeight() int {
	if m.height > 0 {
		return m.height - ui.AppBorderOverheadY
	}
	return 0
}

// borderWidth returns the width to use for internal borders
func (m *Model) borderWidth() int {
	return m.contentWidth()
}

// sessionMaxVisibleItems returns the actual number of session items that can be shown
// based on window height, accounting for fixed UI elements
func (m *Model) sessionMaxVisibleItems() int {
	contentH := m.contentHeight()
	if contentH > 0 {
		// Reserve: header(1) + header border(1) + footer border(1) + message(1) + statusline(1) + help(2) = 7 lines
		availableForContent := contentH - 7
		if availableForContent > 0 {
			return availableForContent
		}
	}
	// Fallback when height unknown
	return 10
}

// projectMaxVisibleItems returns the actual number of items that can be shown
// based on window height, matching the View's calculation
func (m *Model) projectMaxVisibleItems() int {
	contentH := m.contentHeight()
	if contentH > 0 {
		// Reserve: header(1) + header border(1) + footer border(1) + statusline(1) + help(1) = 5 lines
		availableForContent := contentH - 5
		if availableForContent > 0 {
			return availableForContent
		}
	}
	// Fallback when height unknown
	return 10
}

// updateProjectScrollOffset adjusts scroll offset to keep cursor visible in project list
func (m *Model) updateProjectScrollOffset() {
	maxVisible := m.projectMaxVisibleItems()
	// If cursor is above visible area, scroll up
	if m.projectCursor < m.projectScrollOffset {
		m.projectScrollOffset = m.projectCursor
	}
	// If cursor is below visible area, scroll down
	if m.projectCursor >= m.projectScrollOffset+maxVisible {
		m.projectScrollOffset = m.projectCursor - maxVisible + 1
	}
	// Ensure scroll offset is not negative
	if m.projectScrollOffset < 0 {
		m.projectScrollOffset = 0
	}
}

// fuzzyMatch checks if the pattern matches the text (case-insensitive, substring match)
func fuzzyMatch(text, pattern string) bool {
	textLower := strings.ToLower(text)
	return strings.Contains(textLower, pattern)
}

// isCursorValid returns true if cursor points to a valid item
func (m *Model) isCursorValid() bool {
	return m.cursor >= 0 && m.cursor < len(m.items)
}

// getTargetName returns the tmux target name for the given item
func (m *Model) getTargetName(item Item) string {
	if item.IsSession {
		return m.sessions[item.SessionIndex].Name
	}
	session := m.sessions[item.SessionIndex]
	window := session.Windows[item.WindowIndex]
	return fmt.Sprintf("%s:%d", session.Name, window.Index)
}

// setError sets an error message on the model
func (m *Model) setError(format string, args ...any) {
	m.message = fmt.Sprintf(format, args...)
	m.messageIsError = true
}

// sanitizeSessionName converts a path to a valid tmux session name
// Dots and colons have special meaning in tmux target syntax (window.pane, session:window)
// Spaces cause issues with shell commands
func sanitizeSessionName(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		".", "-",
		":", "-",
		" ", "-",
	)
	return replacer.Replace(name)
}

// View implements tea.Model
func (m Model) View() string {
	if m.mode == ModePickDirectory || m.mode == ModeConfirmRemoveFolder {
		return m.viewPickDirectory()
	}
	if m.mode == ModeCloneRepo {
		return m.viewCloneRepo()
	}
	return m.viewSessionList()
}

// viewPickDirectory renders the directory picker view
func (m Model) viewPickDirectory() string {
	var b strings.Builder
	usedLines := 0

	// Header - always show "Select directory", append filter if active
	if m.projectFilter != "" {
		b.WriteString(ui.HeaderStyle.Render("Select directory"))
		b.WriteString("  ")
		b.WriteString(ui.FilterStyle.Render(m.projectFilter))
	} else {
		b.WriteString(ui.HeaderStyle.Render("Select directory"))
	}
	b.WriteString("\n")
	usedLines++

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")
	usedLines++

	// Use shared helper for consistent visible item calculation
	maxItems := m.projectMaxVisibleItems()

	// Directory list (only visible items)
	endIdx := m.projectScrollOffset + maxItems
	if endIdx > len(m.projectFiltered) {
		endIdx = len(m.projectFiltered)
	}
	visibleCount := endIdx - m.projectScrollOffset

	// Get scrollbar characters for each line
	scrollbar := ui.ScrollbarChars(len(m.projectFiltered), maxItems, m.projectScrollOffset, visibleCount)

	contentLines := 0
	for i := m.projectScrollOffset; i < endIdx; i++ {
		fullPath := m.projectFiltered[i]
		displayPath := m.extractDisplayPath(fullPath)
		selected := i == m.projectCursor
		lineIdx := i - m.projectScrollOffset

		// Scrollbar on the left
		if lineIdx < len(scrollbar) {
			b.WriteString(scrollbar[lineIdx])
			b.WriteString(" ")
		}

		if selected {
			b.WriteString(ui.FilterStyle.Render(displayPath))
		} else {
			b.WriteString(displayPath)
		}
		b.WriteString("\n")
		contentLines++
	}

	// Empty state
	if len(m.projectFiltered) == 0 {
		if m.projectFilter != "" {
			b.WriteString("  No directories matching filter\n")
		} else {
			b.WriteString("  No directories found\n")
		}
		contentLines++
	}
	usedLines += contentLines

	// Add padding to push footer to bottom
	// Footer = border (1) + statusline (1) + help line (1) = 3 lines
	// Add 1 more for message line when in confirmation mode
	footerLines := 3
	if m.mode == ModeConfirmRemoveFolder {
		footerLines = 4
	}
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - usedLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

	// Message line (only for confirmation mode)
	if m.mode == ModeConfirmRemoveFolder && m.message != "" {
		if m.messageIsError {
			b.WriteString(ui.ErrorMessageStyle.Render(m.message))
		} else {
			b.WriteString(ui.MessageStyle.Render(m.message))
		}
		b.WriteString("\n")
	}

	// Statusline (directory counts)
	var statusline string
	if m.projectFilter != "" {
		statusline = fmt.Sprintf("%d/%d directories", len(m.projectFiltered), len(m.projectDirs))
	} else {
		statusline = fmt.Sprintf("%d directories", len(m.projectDirs))
	}
	b.WriteString(ui.StatuslineStyle.Render(statusline))
	b.WriteString("\n")

	// Help line
	switch m.mode {
	case ModeConfirmRemoveFolder:
		b.WriteString(ui.FooterStyle.Render(ui.HelpConfirmRemoveFolder()))
	default:
		if m.projectFilter != "" {
			b.WriteString(ui.FooterStyle.Render(ui.HelpFiltering()))
		} else {
			b.WriteString(ui.FooterStyle.Render(ui.HelpPickDirectory()))
		}
	}
	return ui.AppStyle.Render(b.String())
}

// viewCloneRepo renders the clone repository view
func (m Model) viewCloneRepo() string {
	var b strings.Builder
	usedLines := 0

	// Header
	if m.cloneSuccess {
		b.WriteString(ui.HeaderStyle.Render("Clone complete"))
	} else if m.cloneFilter != "" {
		b.WriteString(ui.HeaderStyle.Render("Clone repository"))
		b.WriteString("  ")
		b.WriteString(ui.FilterStyle.Render(m.cloneFilter))
	} else {
		b.WriteString(ui.HeaderStyle.Render("Clone repository"))
	}
	b.WriteString("\n")
	usedLines++

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")
	usedLines++

	// Content area
	contentLines := 0

	if m.cloneSuccess {
		// Show success message with session info
		b.WriteString(fmt.Sprintf("  Cloned: %s\n", m.cloneCloningRepo))
		contentLines++
		b.WriteString(fmt.Sprintf("  Session: %s\n", m.cloneSuccessSession))
		contentLines++
		b.WriteString("\n")
		contentLines++
		b.WriteString("  Switch to the new session?\n")
		contentLines++
	} else if m.cloneLoading {
		b.WriteString("  Fetching available repositories...\n")
		contentLines++
	} else if m.cloneCloning {
		b.WriteString(fmt.Sprintf("  Cloning %s...\n", m.cloneCloningRepo))
		contentLines++
	} else if m.cloneError != "" {
		b.WriteString(ui.ErrorMessageStyle.Render("  "+m.cloneError) + "\n")
		contentLines++
	} else if len(m.cloneRepos) == 0 {
		if m.cloneFilter != "" {
			b.WriteString("  No repositories matching filter\n")
		} else {
			b.WriteString("  No repositories available to clone\n")
		}
		contentLines++
	} else {
		// Repository list
		maxItems := m.cloneMaxVisibleItems()
		endIdx := m.cloneScrollOffset + maxItems
		if endIdx > len(m.cloneRepos) {
			endIdx = len(m.cloneRepos)
		}
		visibleCount := endIdx - m.cloneScrollOffset

		scrollbar := ui.ScrollbarChars(len(m.cloneRepos), maxItems, m.cloneScrollOffset, visibleCount)

		for i := m.cloneScrollOffset; i < endIdx; i++ {
			repo := m.cloneRepos[i]
			selected := i == m.cloneCursor
			lineIdx := i - m.cloneScrollOffset

			if lineIdx < len(scrollbar) {
				b.WriteString(scrollbar[lineIdx])
				b.WriteString(" ")
			}

			if selected {
				b.WriteString(ui.FilterStyle.Render(repo))
			} else {
				b.WriteString(repo)
			}
			b.WriteString("\n")
			contentLines++
		}
	}
	usedLines += contentLines

	// Padding to push footer to bottom
	footerLines := 3
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - usedLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

	// Statusline
	var statusline string
	if m.cloneSuccess {
		statusline = "Clone successful"
	} else if m.cloneLoading {
		statusline = "Loading..."
	} else if m.cloneCloning {
		statusline = "Cloning..."
	} else if m.cloneFilter != "" {
		statusline = fmt.Sprintf("%d/%d repositories", len(m.cloneRepos), len(m.cloneReposAll))
	} else {
		statusline = fmt.Sprintf("%d repositories", len(m.cloneReposAll))
	}
	b.WriteString(ui.StatuslineStyle.Render(statusline))
	b.WriteString("\n")

	// Help line
	if m.cloneSuccess {
		b.WriteString(ui.FooterStyle.Render(ui.HelpCloneSuccess()))
	} else if m.cloneLoading || m.cloneCloning {
		b.WriteString(ui.FooterStyle.Render(ui.HelpCloneRepoLoading()))
	} else if m.cloneFilter != "" {
		b.WriteString(ui.FooterStyle.Render(ui.HelpFiltering()))
	} else {
		b.WriteString(ui.FooterStyle.Render(ui.HelpCloneRepo()))
	}

	return ui.AppStyle.Render(b.String())
}

// viewSessionList renders the main session list view
func (m Model) viewSessionList() string {
	var b strings.Builder
	usedLines := 0

	// Header with optional filter
	if m.filter != "" {
		b.WriteString(ui.HeaderStyle.Render("tsm"))
		b.WriteString("  ")
		b.WriteString(ui.FilterStyle.Render(m.filter))
	} else {
		b.WriteString(ui.HeaderStyle.Render("tsm"))
	}
	b.WriteString("\n")
	usedLines++

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")
	usedLines++

	// Session list (only visible items)
	maxVisible := m.sessionMaxVisibleItems()
	endIdx := m.scrollOffset + maxVisible
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}
	visibleCount := endIdx - m.scrollOffset

	// Get scrollbar characters for each line
	scrollbar := ui.ScrollbarChars(len(m.items), maxVisible, m.scrollOffset, visibleCount)

	// Calculate session numbers (count sessions before visible area)
	sessionNum := 0
	for i := 0; i < m.scrollOffset && i < len(m.items); i++ {
		if m.items[i].IsSession {
			sessionNum++
		}
	}

	contentLines := 0
	for i := m.scrollOffset; i < endIdx; i++ {
		item := m.items[i]
		selected := i == m.cursor
		lineIdx := i - m.scrollOffset

		// Scrollbar on the left
		if lineIdx < len(scrollbar) {
			b.WriteString(scrollbar[lineIdx])
		}

		if item.IsSession {
			session := m.sessions[item.SessionIndex]
			sessionNum++
			isFirst := sessionNum == 1
			b.WriteString(m.renderSessionWithLabel(session, sessionNum, isFirst, selected))
		} else {
			session := m.sessions[item.SessionIndex]
			window := session.Windows[item.WindowIndex]
			b.WriteString(m.renderWindow(window, selected))
		}
		b.WriteString("\n")
		contentLines++
	}

	// Empty state
	if len(m.items) == 0 {
		if m.filter != "" {
			b.WriteString("  No sessions matching filter\n")
		} else {
			b.WriteString("  No other sessions available\n")
		}
		contentLines++
	}
	usedLines += contentLines

	// Message line content (only rendered when there's content)
	var messageContent string
	if m.message != "" {
		if m.messageIsError {
			messageContent = ui.ErrorMessageStyle.Render(m.message)
		} else {
			messageContent = ui.MessageStyle.Render(m.message)
		}
	} else if m.mode == ModeCreate {
		messageContent = ui.InputPromptStyle.Render(" New session: ") + m.input.View()
	}

	// Add padding to push footer to bottom
	// Footer: border (1) + message (1) + statusline (1) + help (2 lines in normal mode)
	footerLines := 5
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - usedLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

	// Message line (always present, may be empty)
	b.WriteString(messageContent)
	b.WriteString("\n")

	// Statusline (session counts)
	var statusline string
	if m.filter != "" {
		// Count visible sessions (items that are sessions, not windows)
		visibleSessions := 0
		for _, item := range m.items {
			if item.IsSession {
				visibleSessions++
			}
		}
		statusline = fmt.Sprintf("%d/%d sessions", visibleSessions, len(m.sessions))
	} else {
		statusline = fmt.Sprintf("%d sessions", len(m.sessions))
	}
	b.WriteString(ui.StatuslineStyle.Render(statusline))
	b.WriteString("\n")

	// Help line
	switch m.mode {
	case ModeNormal:
		if m.filter != "" {
			b.WriteString(ui.FooterStyle.Render(ui.HelpFiltering()))
		} else {
			b.WriteString(ui.FooterStyle.Render(ui.HelpNormal()))
		}
	case ModeConfirmKill:
		b.WriteString(ui.FooterStyle.Render(ui.HelpConfirmKill()))
	case ModeCreate:
		b.WriteString(ui.FooterStyle.Render(ui.HelpCreate()))
	}

	return ui.AppStyle.Render(b.String())
}

func (m Model) renderSessionWithLabel(session tmux.Session, num int, isFirst bool, selected bool) string {
	// Build the row with fixed-width columns
	var b strings.Builder

	// Number label
	label := fmt.Sprintf("%d", num)
	if selected {
		b.WriteString(ui.IndexSelectedStyle.Render(label))
	} else {
		b.WriteString(ui.IndexStyle.Render(label))
	}
	b.WriteString(" ")

	// Last session icon (fixed width column)
	if isFirst {
		if selected {
			b.WriteString(ui.LastIconSelected)
		} else {
			b.WriteString(ui.LastIcon)
		}
	} else {
		b.WriteString(" ")
	}
	b.WriteString(" ")

	// Expand icon
	if session.Expanded {
		if selected {
			b.WriteString(ui.ExpandedIconSelected)
		} else {
			b.WriteString(ui.ExpandedIcon)
		}
	} else {
		if selected {
			b.WriteString(ui.CollapsedIconSelected)
		} else {
			b.WriteString(ui.CollapsedIcon)
		}
	}
	b.WriteString(" ")

	// Session name (padded to max width)
	namePadded := fmt.Sprintf("%-*s", m.maxNameWidth, session.Name)
	if selected {
		b.WriteString(ui.SessionNameSelectedStyle.Render(namePadded))
	} else {
		b.WriteString(namePadded)
	}
	b.WriteString("  ")

	// Time ago (fixed width 8)
	timeAgo := formatTimeAgo(session.LastActivity)
	timePadded := fmt.Sprintf("%-8s", timeAgo)
	if selected {
		b.WriteString(ui.TimeSelectedStyle.Render(timePadded))
	} else {
		b.WriteString(ui.TimeStyle.Render(timePadded))
	}

	// Git status (fixed width column)
	if m.maxGitStatusWidth > 0 {
		b.WriteString(" ")
		if status, ok := m.gitStatuses[session.Name]; ok {
			formatted := ui.FormatGitStatus(status.Dirty, status.Ahead, status.Behind)
			actualWidth := ui.GitStatusWidth(status.Dirty, status.Ahead, status.Behind)
			b.WriteString(formatted)
			// Pad to max width
			if actualWidth < m.maxGitStatusWidth {
				b.WriteString(strings.Repeat(" ", m.maxGitStatusWidth-actualWidth))
			}
		} else {
			// Empty placeholder for alignment
			b.WriteString(strings.Repeat(" ", m.maxGitStatusWidth))
		}
	}

	// Claude status
	if status, ok := m.claudeStatuses[session.Name]; ok {
		b.WriteString(" ")
		b.WriteString(ui.FormatClaudeStatus(status.State, m.animationFrame))
	}

	return ui.SessionStyle.Render(b.String())
}

func (m Model) renderWindow(window tmux.Window, selected bool) string {
	var b strings.Builder

	// Window index and name
	windowText := fmt.Sprintf("%d: %s", window.Index, window.Name)
	if selected {
		b.WriteString(ui.WindowNameSelectedStyle.Render(windowText))
	} else {
		b.WriteString(windowText)
	}

	return ui.WindowStyle.Render(b.String())
}

func formatTimeAgo(t time.Time) string {
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
