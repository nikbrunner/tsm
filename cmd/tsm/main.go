package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nikbrunner/tsm/internal/config"
	"github.com/nikbrunner/tsm/internal/model"
	"github.com/nikbrunner/tsm/internal/tmux"
)

func main() {
	// Ensure HOME is set (required for config paths)
	if os.Getenv("HOME") == "" {
		fmt.Println("Error: HOME environment variable not set")
		os.Exit(1)
	}

	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := config.Init(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created config file at %s\n", config.Path())
			return
		case "bookmark":
			if len(os.Args) < 3 {
				fmt.Println("Usage: tsm bookmark <N>")
				fmt.Println("Opens bookmark at slot N (1-9)")
				os.Exit(1)
			}
			if err := runBookmark(os.Args[2]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "tmux-bindings":
			if err := printTmuxBindings(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Usage: tsm [init | bookmark <N> | tmux-bindings]")
			os.Exit(1)
		}
	}

	// Check if running inside tmux
	if os.Getenv("TMUX") == "" {
		fmt.Println("Error: tsm must be run from within tmux")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get current session to exclude from list
	currentSession, err := tmux.CurrentSession()
	if err != nil {
		fmt.Printf("Error getting current session: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run the TUI
	m := model.New(currentSession, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// runBookmark opens the bookmark at slot N (1-9)
func runBookmark(slotStr string) error {
	slot, err := strconv.Atoi(slotStr)
	if err != nil || slot < 1 || slot > 9 {
		return fmt.Errorf("invalid slot: %s (must be 1-9)", slotStr)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	idx := slot - 1
	if idx >= len(cfg.Bookmarks) {
		return fmt.Errorf("no bookmark at slot %d", slot)
	}

	bookmark := cfg.Bookmarks[idx]
	sessionName := filepath.Base(bookmark.Path)

	// Create session if it doesn't exist
	if !tmux.SessionExists(sessionName) {
		if err := tmux.CreateSession(sessionName, bookmark.Path); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Apply layout if configured
		if cfg.Layout != "" && cfg.LayoutDir != "" {
			layoutPath := filepath.Join(cfg.LayoutDir, cfg.Layout+".sh")
			if _, err := os.Stat(layoutPath); err == nil {
				// Run layout script with session name as argument
				cmd := exec.Command(layoutPath, sessionName)
				cmd.Dir = bookmark.Path
				_ = cmd.Run() // Ignore layout errors
			}
		}
	}

	// Switch to the session
	return tmux.SwitchClient(sessionName)
}

// printTmuxBindings outputs tmux bind commands for configured bookmarks
func printTmuxBindings() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Bookmarks) == 0 {
		fmt.Println("# No bookmarks configured")
		fmt.Println("# Add bookmarks to ~/.config/tsm/config.yml:")
		fmt.Println("# bookmarks:")
		fmt.Println("#   - path: ~/repos/my-project")
		return nil
	}

	fmt.Println("# tsm bookmark bindings")
	fmt.Println("# Add to your tmux.conf or source with: run-shell \"tsm tmux-bindings | tmux source-stdin\"")
	for i := range cfg.Bookmarks {
		slot := i + 1
		if slot > 9 {
			break // Only slots 1-9
		}
		fmt.Printf("bind -n M-%d run-shell \"tsm bookmark %d\"\n", slot, slot)
	}

	return nil
}
