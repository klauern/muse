package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/hooks"
	"github.com/urfave/cli/v2"
)

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
	gitDir, err := hooks.FindGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		fmt.Println("prepare-commit-msg hook is not installed")
	} else {
		fmt.Println("prepare-commit-msg hook is installed")
	}

	fmt.Printf("Hook configuration: DryRun=%t Type=%s\n", config.Hook.DryRun, config.Hook.Type)

	fmt.Println("Status check completed")
	return nil
}
