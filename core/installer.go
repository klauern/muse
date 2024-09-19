package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
	// Join the arguments into a single string
	argsStr := ""
	for _, arg := range args {
		argsStr += fmt.Sprintf(" \"%s\"", arg)
	}

	// Construct the hook content
	hookContent := fmt.Sprintf(`#!/bin/sh
exec < /dev/tty
%s/%s%s
`, binaryPath, binaryName, argsStr)

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
	args := []string{"prepare-commit-msg", "$1", "$2", "$3"} // Example arguments

	hookContent := generateHookContent(binaryPath, binaryName, args)

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
