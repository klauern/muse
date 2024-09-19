package hook

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/pre-commit-llm/config"
)

type Installer struct {
	config *config.Config
}

func NewInstaller(config *config.Config) *Installer {
	return &Installer{config: config}
}

func (i *Installer) Install() error {
	gitDir, err := FindGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		return fmt.Errorf("prepare-commit-msg hook already exists")
	}

	// Create the hook file
	f, err := os.Create(hookPath)
	if err != nil {
		return fmt.Errorf("failed to create hook file: %w", err)
	}
	defer f.Close()

	// Write the hook content
	hookContent := `#!/bin/sh
exec < /dev/tty
muse prepare-commit-msg "$1" "$2" "$3"
`
	if _, err := f.WriteString(hookContent); err != nil {
		return fmt.Errorf("failed to write hook content: %w", err)
	}

	// Make the hook executable
	if err := os.Chmod(hookPath, 0o755); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	fmt.Println("prepare-commit-msg hook installed successfully")
	return nil
}

func FindGitDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return filepath.Join(dir, ".git"), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}
