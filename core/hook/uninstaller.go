package hook

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/pre-commit-llm/config"
)

func UninstallHook(config *config.Config) error {
	gitDir, err := FindGitDir()
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
	fmt.Println("Hook uninstalled successfully")
	return nil
}
