package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yourusername/muse"
)

func NewStatusCmd(config *muse.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check the status of the prepare-commit-msg hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkStatus(config)
		},
	}
}

func checkStatus(config *muse.Config) error {
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

	fmt.Printf("Hook configuration: Enabled=%v, Type=%s\n", config.HookConfig.Enabled, config.HookConfig.Type)

	return nil
}
