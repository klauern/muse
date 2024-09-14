package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/klauern/pre-commit-llm"
)

func NewUninstallCmd(config *muse.Config) *cli.Command {
	return &cli.Command{
		Name:  "uninstall",
		Usage: "Uninstall the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			return uninstallHook(config)
		},
	}
}

func uninstallHook(config *muse.Config) error {
	gitDir, err := findGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")
	
	// Check if hook exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("prepare-commit-msg hook does not exist")
	}

	// Remove the hook file
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove hook file: %w", err)
	}

	fmt.Println("prepare-commit-msg hook uninstalled successfully")
	return nil
}
