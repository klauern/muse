package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/urfave/cli/v2"
)

func NewInstallCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			return installHook(config)
		},
	}
}

func installHook(config *config.Config) error {
	gitDir, err := findGitDir()
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
"$GOPATH/bin/muse" prepare-commit-msg "$@"
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

func findGitDir() (string, error) {
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
