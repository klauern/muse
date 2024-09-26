package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/klauern/muse/config"
)

type Installer struct {
	config *config.Config
}

func NewInstaller(config *config.Config) *Installer {
	return &Installer{config: config}
}

const (
	hookStartMarker = "# BEGIN MUSE HOOK"
	hookEndMarker   = "# END MUSE HOOK"
)

func addOrUpdateHookContent(hookPath, hookContent string) error {
	var existingContent []byte
	var err error

	// Check if the hook file exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		// Case 2: File doesn't exist, create it with shebang and hook content
		fullContent := "#!/bin/sh\n\n" + hookContent
		return os.WriteFile(hookPath, []byte(fullContent), 0o755)
	}

	// File exists, read its content
	existingContent, err = os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("failed to read hook file: %w", err)
	}

	// Remove any existing MUSE hook content
	re := regexp.MustCompile(fmt.Sprintf("(?s)%s.*?%s", regexp.QuoteMeta(hookStartMarker), regexp.QuoteMeta(hookEndMarker)))
	updatedContent := re.ReplaceAllString(string(existingContent), "")

	// Case 1: File already has content (e.g., lefthook)
	if len(strings.TrimSpace(updatedContent)) > 0 {
		// Append the new hook content without shebang
		updatedContent = strings.TrimSpace(updatedContent) + "\n\n" + hookContent
	} else {
		// Case 2: File is empty or only contained MUSE hook
		updatedContent = "#!/bin/sh\n\n" + hookContent
	}

	// Write updated content back to the file
	return os.WriteFile(hookPath, []byte(updatedContent), 0o755)
}

func removeHookContent(hookPath string) error {
	// Read existing hook file content
	existingContent, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("failed to read hook file: %w", err)
	}

	// Remove content between markers
	re := regexp.MustCompile(fmt.Sprintf("(?s)%s.*?%s", regexp.QuoteMeta(hookStartMarker), regexp.QuoteMeta(hookEndMarker)))
	updatedContent := re.ReplaceAllString(string(existingContent), "")

	// Write updated content back to the file
	if err := os.WriteFile(hookPath, []byte(updatedContent), 0o755); err != nil {
		return fmt.Errorf("failed to write hook content: %w", err)
	}

	return nil
}

func generateHookContent(binaryPath, binaryName string, args []string) string {
	// Construct the hook content
	hookContent := fmt.Sprintf(`%s
# Save the original arguments
COMMIT_MSG_FILE="$1"
COMMIT_SOURCE="$2"
SHA1="$3"

# Check if verbose flag is set
if [ "$MUSE_VERBOSE" = "true" ]; then
    VERBOSE_FLAG="--verbose"
else
    VERBOSE_FLAG=""
fi

# Execute the binary with the saved arguments
%s/%s prepare-commit-msg $VERBOSE_FLAG "$COMMIT_MSG_FILE" "$COMMIT_SOURCE" "$SHA1"
%s
`, hookStartMarker, binaryPath, binaryName, hookEndMarker)

	return hookContent
}

func (i *Installer) Install() error {
	gitDir, err := FindGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	// Get the path of the currently running executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	binaryPath := filepath.Dir(exePath)
	binaryName := filepath.Base(exePath)
	args := []string{"prepare-commit-msg", "$1", "$2", "$3"}

	hookContent := generateHookContent(binaryPath, binaryName, args)

	fmt.Printf("Installing prepare-commit-msg hook... at %s\n", hookPath)
	if err := addOrUpdateHookContent(hookPath, hookContent); err != nil {
		return fmt.Errorf("failed to add or update hook content: %w", err)
	}

	fmt.Println("prepare-commit-msg hook installed successfully")
	return nil
}

func (i *Installer) Uninstall() error {
	gitDir, err := FindGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	if _, err := os.Stat(hookPath); err == nil {
		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("failed to remove hook: %w", err)
		}
		fmt.Println("prepare-commit-msg hook uninstalled successfully")
	} else if os.IsNotExist(err) {
		fmt.Println("prepare-commit-msg hook does not exist")
	} else {
		return fmt.Errorf("failed to check hook existence: %w", err)
	}

	return nil
}
