package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/urfave/cli/v2"
)

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
			return "", fmt.Errorf("not a git repository (or any parent up to root)")
		}
		dir = parent
	}
}

func NewStatusCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Check the status of the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			return checkStatus(config)
		},
	}
}

func checkStatus(config *config.Config) error {
	gitDir, err := findGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		fmt.Println("prepare-commit-msg hook is not installed")
	} else {
		fmt.Println("prepare-commit-msg hook is installed")
	}

	fmt.Printf("Hook configuration: Enabled=%v, Type=%s\n", config.Hook.Enabled, config.Hook.Type)

	return nil
}
