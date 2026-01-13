package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration options for tsm
type Config struct {
	// Layout script name to apply when creating new sessions
	Layout string `yaml:"layout"`

	// Directory containing layout scripts
	LayoutDir string `yaml:"layout_dir"`

	// Enable Claude Code status integration
	ClaudeStatusEnabled bool `yaml:"claude_status_enabled"`

	// Enable git status indicator in session list
	GitStatusEnabled bool `yaml:"git_status_enabled"`

	// Directory for status cache files
	CacheDir string `yaml:"cache_dir"`

	// Base directories for project picker (C-p) - supports multiple paths
	ProjectDirs []string `yaml:"project_dirs"`

	// Scan depth for project directories (default: 2 for owner/repo structure)
	ProjectDepth int `yaml:"project_depth"`

	// Default directory for new sessions created with C-n
	DefaultSessionDir string `yaml:"default_session_dir"`

	// Lazygit popup dimensions
	LazygitPopup PopupConfig `yaml:"lazygit_popup"`

	// Quick-access session bookmarks (slots 1-9, maps to M-1 through M-9)
	Bookmarks []Bookmark `yaml:"bookmarks,omitempty"`
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

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() Config {
	home := os.Getenv("HOME")
	return Config{
		Layout:              "",
		LayoutDir:           filepath.Join(home, ".config", "tmux", "layouts"),
		ClaudeStatusEnabled: false,
		GitStatusEnabled:    false,
		CacheDir:            filepath.Join(home, ".cache", "tsm"),
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
	return filepath.Join(home, ".config", "tsm", "config.yml")
}

// BookmarksPath returns the path to the separate bookmarks file
func BookmarksPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "tsm", "bookmarks.yml")
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
	if os.Getenv("TMUX_SESSION_PICKER_GIT_STATUS") == "1" {
		cfg.GitStatusEnabled = true
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

	// Write default config with comments
	content := `# tsm configuration
# Environment variables override these settings

# Layout script name to apply when creating new sessions
# layout: ide

# Directory containing layout scripts
# layout_dir: ~/.config/tmux/layouts

# Enable Claude Code status integration
# claude_status_enabled: false

# Enable git status indicator (shows dirty/ahead/behind for repos)
# git_status_enabled: false

# Directory for status cache files
# cache_dir: ~/.cache/tsm

# Base directories for project picker (C-p)
# Supports multiple paths - all will be scanned
# project_dirs:
#   - ~/repos
# Example with multiple paths:
# project_dirs:
#   - ~/repos
#   - ~/work
#   - ~/personal

# Scan depth for project directories (2 = owner/repo structure)
# project_depth: 2

# Default directory for new sessions created with C-n
# default_session_dir: ~

# Lazygit popup dimensions (C-g)
# lazygit_popup:
#   width: 90%
#   height: 90%

# Quick-access session bookmarks (slots 1-9, maps to M-1 through M-9)
# Use 'tsm tmux-bindings' to generate tmux keybindings
# Note: Bookmarks are stored separately in ~/.config/tsm/bookmarks.yml
# to preserve comments in this file when bookmarks are modified via the TUI.
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveBookmarks writes bookmarks to the separate bookmarks file
// This preserves comments in the main config.yml
func (cfg *Config) SaveBookmarks() error {
	bf := BookmarksFile{Bookmarks: cfg.Bookmarks}
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
	if len(path) > 0 && path[0] == '~' {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[1:])
	}
	return path
}
