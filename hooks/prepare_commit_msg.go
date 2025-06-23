package hooks

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/internal/fileops"
	"github.com/klauern/muse/internal/git"
	"github.com/klauern/muse/internal/userinput"
	"github.com/klauern/muse/llm"
)

type PrepareCommitMsgHook interface {
	Run(commitMsgFile string, commitSource string, sha1 string) error
}

type LLMHook struct {
	Generator llm.Generator
	Config    *config.Config
}

func (h *LLMHook) Run(commitMsgFile string, commitSource string, sha1 string) error {
	// Always get the staged changes
	diff, err := getGitDiff()
	if err != nil {
		slog.Error("Failed to get git diff", "error", err)
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Get the commit style from the configuration
	commitStyle := h.Config.Hook.CommitStyle

	// Generate the commit message
	ctx := context.Background()
	fmt.Println("Generating commit message")
	message, err := h.Generator.Generate(ctx, diff, commitStyle)
	if err != nil {
		slog.Error("Failed to generate commit message", "error", err)
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Check if dry run mode is enabled
	if h.Config.Hook.DryRun {
		fmt.Println("Dry run mode: Generated commit message:")
		fmt.Println(message)
		return nil
	}

	// Check if preview mode is enabled
	if h.Config.Hook.Preview {
		fmt.Println("Preview mode: Generated commit message:")
		fmt.Println(message)

		// Use secure input handler with timeout and validation
		inputHandler := userinput.NewSecureInputHandler()
		accepted, err := inputHandler.PromptYesNo(ctx, "Do you want to use this commit message? (y/n): ")
		if err != nil {
			slog.Error("Failed to read user input", "error", err)
			return fmt.Errorf("failed to read user input: %w", err)
		}

		if !accepted {
			slog.Info("User rejected the generated commit message")
			return fmt.Errorf("user rejected the generated commit message")
		}
	}

	// Debug: Log the message content and length
	slog.Debug("Generated commit message", "message", message, "length", len(message))
	
	// Check if message is empty or whitespace-only
	trimmedMessage := strings.TrimSpace(message)
	if len(trimmedMessage) == 0 {
		slog.Error("Generated commit message is empty or whitespace-only", "original", message)
		return fmt.Errorf("generated commit message is empty or whitespace-only")
	}

	// Write the generated message to the commit message file atomically
	if err := fileops.SafeWriteFile(commitMsgFile, []byte(message), 0o644); err != nil {
		slog.Error("Failed to write commit message", "error", err)
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	slog.Info("Commit message successfully generated and saved", "message", message)

	return nil
}

// getGitDiff safely retrieves staged changes using our secure Git operations
func getGitDiff() (string, error) {
	gitOps, err := git.NewGitOperations("")
	if err != nil {
		slog.Error("Failed to initialize git operations", "error", err)
		return "", fmt.Errorf("failed to initialize git operations: %w", err)
	}

	diff, err := gitOps.GetStagedDiff()
	if err != nil {
		slog.Error("Failed to get staged diff", "error", err)
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	return diff, nil
}

func NewHook(cfg *config.Config) (PrepareCommitMsgHook, error) {
	generator, err := llm.NewCommitMessageGenerator(cfg)
	if err != nil {
		slog.Error("Failed to create commit message generator", "error", err)
		return nil, fmt.Errorf("failed to create commit message generator: %w", err)
	}
	return &LLMHook{Generator: generator, Config: cfg}, nil
}
