package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Appearance represents the terminal color scheme mode
type Appearance string

const (
	AppearanceDark  Appearance = "dark"
	AppearanceLight Appearance = "light"
)

// Config holds all configuration options for helm
type Config struct {
	// Terminal appearance mode ("light" or "dark") - controls color adaptation
	Appearance Appearance `yaml:"appearance"`

	// Black Atom theme key (e.g. "black-atom-jpn-koyo-yoru").
	// When set, the theme's colors are used and its own appearance wins
	// over the appearance key. Empty = terminal-native colors (ANSI-16 +
	// reverse video). Env override: HELM_THEME.
	Theme string `yaml:"theme,omitempty"`

	// Layout script name to apply when creating new sessions
	Layout string `yaml:"layout"`

	// Directory containing layout scripts
	LayoutDir string `yaml:"layout_dir"`

	// Enable layout scripts feature (when disabled, layouts won't be auto-applied)
	EnableLayouts bool `yaml:"enable_layouts"`

	// Enable Claude Code status integration
	ClaudeStatusEnabled bool `yaml:"claude_status_enabled"`

	// Enable Pi status integration
	PiStatusEnabled bool `yaml:"pi_status_enabled"`

	// Enable git status indicator in session list
	GitStatusEnabled bool `yaml:"git_status_enabled"`

	// Directory for status cache files
	CacheDir string `yaml:"cache_dir"`

	// Base directories for project picker (C-p) - supports multiple paths
	ProjectDirs []string `yaml:"project_dirs"`

	// Scan depth for project directories (default: 2 for owner/repo structure)
	ProjectDepth int `yaml:"project_depth"`

	// Maps git hosts to directory aliases for clone destination paths.
	// Empty string = use path as-is (GitHub style). Omitted hosts use host/ as prefix.
	// Example: {"git.corp.example.com": "corp"} resolves ssh://git@git.corp.example.com:7999/~alice/proj
	// to corp/~alice/proj instead of git.corp.example.com/~alice/proj
	GitProviders map[string]string `yaml:"git_providers,omitempty"`

	// Default directory for new sessions created with C-n
	DefaultSessionDir string `yaml:"default_session_dir"`

	// Lazygit popup dimensions
	LazygitPopup PopupConfig `yaml:"lazygit_popup"`

	// Quick-access session bookmarks (slots 1-9, maps to M-1 through M-9)
	Bookmarks []Bookmark `yaml:"bookmarks,omitempty"`

	// Command to run on each dirty repo via 'helm repos dirty --walk'
	// Use {} as placeholder for the repo path, e.g. "lazygit -p {}"
	DirtyWalkthroughCommand string `yaml:"dirty_walkthrough_command,omitempty"`

	// Repositories to ensure are cloned (used by helm setup)
	EnsureCloned []EnsureClonedEntry `yaml:"ensure_cloned,omitempty"`
}

// PopupConfig holds popup dimension settings
type PopupConfig struct {
	Width  string `yaml:"width"`
	Height string `yaml:"height"`
}

// Bookmark represents a quick-access session bookmark
type Bookmark struct {
	Path string `yaml:"path"`
}

// EnsureClonedEntry represents a repository to ensure is cloned.
// Supports both string format (just a URL) and object format (url + post_clone).
type EnsureClonedEntry struct {
	URL       string `yaml:"url"`
	PostClone string `yaml:"post_clone,omitempty"`
}

// UnmarshalYAML allows EnsureClonedEntry to be specified as either a plain string or an object.
func (e *EnsureClonedEntry) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		e.URL = value.Value
		return nil
	}
	type plain EnsureClonedEntry
	return value.Decode((*plain)(e))
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() Config {
	home := os.Getenv("HOME")
	return Config{
		Appearance:          AppearanceDark,
		Layout:              "",
		LayoutDir:           filepath.Join(home, ".config", "tmux", "layouts"),
		EnableLayouts:       false,
		ClaudeStatusEnabled: false,
		PiStatusEnabled:     false,
		GitStatusEnabled:    false,
		CacheDir:            filepath.Join(home, ".cache", AppDirName),
		ProjectDirs:         []string{filepath.Join(home, "repos")},
		ProjectDepth:        2,
		DefaultSessionDir:   home,
		LazygitPopup: PopupConfig{
			Width:  "90%",
			Height: "90%",
		},
	}
}

// Path returns the path to the config file
func Path() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", ConfigDirName(), ConfigFileName)
}

// BookmarksPath returns the path to the separate bookmarks file
func BookmarksPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", ConfigDirName(), BookmarksFileName)
}

// BookmarksFile represents the structure of the bookmarks file
type BookmarksFile struct {
	Bookmarks []Bookmark `yaml:"bookmarks"`
}

// Load reads configuration from file and environment variables.
// Priority: env vars > config file > defaults
func Load() (Config, error) {
	cfg := DefaultConfig()

	// Load from config file if it exists
	configPath := Path()
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Expand ~ in paths
	cfg.LayoutDir = expandPath(cfg.LayoutDir)
	cfg.CacheDir = expandPath(cfg.CacheDir)
	cfg.DefaultSessionDir = expandPath(cfg.DefaultSessionDir)

	// Expand ~ in project directories
	for i, d := range cfg.ProjectDirs {
		cfg.ProjectDirs[i] = expandPath(d)
	}

	// Expand ~ in bookmark paths
	for i := range cfg.Bookmarks {
		cfg.Bookmarks[i].Path = expandPath(cfg.Bookmarks[i].Path)
	}

	// Ensure ProjectDepth is at least 1
	if cfg.ProjectDepth < 1 {
		cfg.ProjectDepth = 2
	}

	// Environment variables override config file
	if val := os.Getenv("TMUX_LAYOUT"); val != "" {
		cfg.Layout = val
	}
	if val := os.Getenv("TMUX_LAYOUTS_DIR"); val != "" {
		cfg.LayoutDir = expandPath(val)
	}
	if os.Getenv("TMUX_SESSION_PICKER_CLAUDE_STATUS") == "1" {
		cfg.ClaudeStatusEnabled = true
	}
	if os.Getenv("TMUX_SESSION_PICKER_PI_STATUS") == "1" {
		cfg.PiStatusEnabled = true
	}
	if os.Getenv("TMUX_SESSION_PICKER_GIT_STATUS") == "1" {
		cfg.GitStatusEnabled = true
	}
	if val := os.Getenv("HELM_THEME"); val != "" {
		cfg.Theme = val
	}

	// Load bookmarks from separate file (takes priority over config.yml bookmarks)
	if bookmarks, err := LoadBookmarks(); err == nil {
		cfg.Bookmarks = bookmarks
	}
	// If no separate bookmarks file exists, keep bookmarks from config.yml (for migration)

	return cfg, nil
}

// LoadBookmarks reads bookmarks from the separate bookmarks file
func LoadBookmarks() ([]Bookmark, error) {
	bookmarksPath := BookmarksPath()
	if _, err := os.Stat(bookmarksPath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(bookmarksPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	var bf BookmarksFile
	if err := yaml.Unmarshal(data, &bf); err != nil {
		return nil, fmt.Errorf("failed to parse bookmarks file: %w", err)
	}

	// Expand ~ in bookmark paths
	for i := range bf.Bookmarks {
		bf.Bookmarks[i].Path = expandPath(bf.Bookmarks[i].Path)
	}

	return bf.Bookmarks, nil
}

// Init creates a new config file with commented defaults
func Init() error {
	configPath := Path()
	configDir := filepath.Dir(configPath)

	// Create directory if needed
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Write minimal config — schema descriptions provide documentation
	content := `# yaml-language-server: $schema=https://raw.githubusercontent.com/black-atom-industries/helm.tmux/main/schema.json
appearance: dark
project_dirs:
  - ~/repos
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveBookmarks writes bookmarks to the separate bookmarks file
// This preserves comments in the main config.yml
func (cfg *Config) SaveBookmarks() error {
	// Contract absolute paths back to ~ before saving
	contracted := make([]Bookmark, len(cfg.Bookmarks))
	for i, b := range cfg.Bookmarks {
		contracted[i] = Bookmark{Path: contractPath(b.Path)}
	}
	bf := BookmarksFile{Bookmarks: contracted}
	data, err := yaml.Marshal(bf)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	// Ensure config directory exists
	bookmarksPath := BookmarksPath()
	if err := os.MkdirAll(filepath.Dir(bookmarksPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(bookmarksPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write bookmarks file: %w", err)
	}
	return nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	// Only expand bare ~ (not ~username)
	if path == "~" || strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[1:])
	}
	return path
}

// contractPath replaces the user's home directory with ~.
// Paths starting with ~ (e.g. ~username) are returned as-is.
func contractPath(path string) string {
	// Already contracted or ~username — leave it alone
	if strings.HasPrefix(path, "~") {
		return path
	}
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	if path == home {
		return "~"
	}
	return path
}

// RepoInfo represents a cloned repository with its location
type RepoInfo struct {
	Name string // "owner/repo"
	Path string // absolute path on disk
}

// ScanForGitRepos walks baseDir recursively, collecting directories that contain .git.
// Stops recursing into directories that are themselves repos.
func ScanForGitRepos(baseDir string) []string {
	var repos []string
	var walk func(dir string, relPath string)
	walk = func(dir string, relPath string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if IsHiddenDir(entry.Name()) {
				continue
			}
			entryRel := entry.Name()
			if relPath != "" {
				entryRel = relPath + "/" + entry.Name()
			}
			entryPath := filepath.Join(dir, entry.Name())
			gitPath := filepath.Join(entryPath, ".git")
			if _, err := os.Stat(gitPath); err == nil {
				// Found a repo — add it, don't recurse into it
				repos = append(repos, entryRel)
				continue
			}
			// Not a repo — recurse deeper
			walk(entryPath, entryRel)
		}
	}
	walk(baseDir, "")
	return repos
}

// IsHiddenDir returns true for VCS and internal metadata directories
// (e.g. .git, .hg, .svn) but false for project directories that happen to
// start with a dot (e.g. .github-private).
func IsHiddenDir(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", ".DS_Store", ".Trash", ".cache", ".local", ".config":
		return true
	}
	return false
}

// ListClonedRepos returns already-cloned repos in owner/repo format.
// Uses recursive walk with .git detection — discovers repos at any depth.
func ListClonedRepos(basePath string) ([]string, error) {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return []string{}, nil
	}
	repos := ScanForGitRepos(basePath)
	sort.Strings(repos)
	return repos, nil
}

// FilterUncloned returns repos from available that are not in cloned
func FilterUncloned(available, cloned []string) []string {
	clonedSet := make(map[string]bool)
	for _, r := range cloned {
		clonedSet[r] = true
	}

	var uncloned []string
	for _, r := range available {
		if !clonedSet[r] {
			uncloned = append(uncloned, r)
		}
	}

	return uncloned
}

// ListAllRepos scans all project directories and returns every cloned repo with its path.
// Deduplicates by name — first occurrence wins.
func ListAllRepos(projectDirs []string) ([]RepoInfo, error) {
	seen := make(map[string]bool)
	var all []RepoInfo

	for _, dir := range projectDirs {
		names, err := ListClonedRepos(dir)
		if err != nil {
			continue
		}
		for _, name := range names {
			if seen[name] {
				continue
			}
			seen[name] = true
			all = append(all, RepoInfo{
				Name: name,
				Path: filepath.Join(dir, name),
			})
		}
	}

	return all, nil
}

// SanitizeSessionName converts a path or identifier to a valid tmux session name.
// Replaces characters with special meaning in tmux target syntax:
//   - "/" (path separator, also used in session:window)
//   - "." (window.pane separator)
//   - ":" (session:window separator)
//   - " " (breaks shell commands)
func SanitizeSessionName(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		".", "-",
		":", "-",
		" ", "-",
	)
	return replacer.Replace(name)
}

// extractRelPath converts an absolute path to a relative form for display or
// session naming. Tries each projectDir in order — the first one that contains
// the path wins. Falls back to the last `depth` components of the path when
// no projectDir matches (or when projectDirs is empty).
func extractRelPath(fullPath string, projectDirs []string, depth int) string {
	for _, projectDir := range projectDirs {
		if rel, err := filepath.Rel(projectDir, fullPath); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	parts := strings.Split(fullPath, string(filepath.Separator))
	if depth > len(parts) {
		depth = len(parts)
	}
	return strings.Join(parts[len(parts)-depth:], "/")
}

// ExtractDisplayPath converts an absolute path to a display path.
// Returns the path relative to the matching projectDir, or the last `depth`
// components if no projectDir matches. Forward slashes are used on all
// platforms for stable display.
func ExtractDisplayPath(fullPath string, projectDirs []string, depth int) string {
	return extractRelPath(fullPath, projectDirs, depth)
}

// ExtractSessionName converts an absolute path to a tmux-safe session name.
// Tries projectDirs first (gives names like "imfusion-websdk-web-ui"), then
// falls back to the last `depth` components for paths outside any projectDir.
//
// This is the single source of truth for session naming — both the TUI
// (project picker, bookmarks view) and the CLI (`helm bookmark N`) must use
// it to ensure the same path always produces the same session name.
func ExtractSessionName(fullPath string, projectDirs []string, depth int) string {
	return SanitizeSessionName(extractRelPath(fullPath, projectDirs, depth))
}
