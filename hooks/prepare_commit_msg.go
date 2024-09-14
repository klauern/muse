package hooks

import (
	"fmt"
	"os"
)

type PrepareCommitMsgHook interface {
	Run(commitMsgFile string, commitSource string, sha1 string) error
}

type DefaultHook struct{}

func (h *DefaultHook) Run(commitMsgFile string, commitSource string, sha1 string) error {
	// Read the commit message
	content, err := os.ReadFile(commitMsgFile)
	if err != nil {
		return fmt.Errorf("failed to read commit message file: %w", err)
	}

	// Modify the commit message (this is just a placeholder)
	modifiedContent := []byte(fmt.Sprintf("Modified: %s", string(content)))

	// Write the modified commit message back to the file
	if err := os.WriteFile(commitMsgFile, modifiedContent, 0644); err != nil {
		return fmt.Errorf("failed to write modified commit message: %w", err)
	}

	return nil
}

func NewHook(hookType string) PrepareCommitMsgHook {
	switch hookType {
	case "default":
		return &DefaultHook{}
	// Add more hook types here as needed
	default:
		return &DefaultHook{}
	}
}
