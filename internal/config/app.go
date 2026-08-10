package config

import "path/filepath"

// App-level constants (not user-configurable)
const (
	AppName = "BLACK ATOM HELM"

	// Directory and file names
	AppDirName        = "helm-tmux"
	ConfigFileName    = "config.yml"
	BookmarksFileName = "bookmarks.yml"
	StatusFileExt     = ".status"
	PiStatusFileExt   = ".pi-status"
)

// ConfigDirName returns the relative path for config files under ~/.config/
func ConfigDirName() string {
	return filepath.Join("black-atom", "helm-tmux")
}
