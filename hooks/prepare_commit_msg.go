package hooks

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/klauern/muse/config"
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
		fmt.Print("Do you want to use this commit message? (y/n): ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			slog.Error("Failed to read user input", "error", err)
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if response != "y" && response != "Y" {
			slog.Error("User rejected the generated commit message")
			return fmt.Errorf("user rejected the generated commit message")
		}
	}

	// Write the generated message to the commit message file
	if err := os.WriteFile(commitMsgFile, []byte(message), 0o644); err != nil {
		slog.Error("Failed to write commit message", "error", err)
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	fmt.Println("Commit message successfully generated and saved.")

	return nil
}

func getGitDiff() (string, error) {
	// Get the staged changes
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to execute git diff command", "error", err)
		return "", err
	}
	return string(output), nil
}

func NewHook(cfg *config.Config) (PrepareCommitMsgHook, error) {
	generator, err := llm.NewCommitMessageGenerator(cfg)
	if err != nil {
		slog.Error("Failed to create commit message generator", "error", err)
		return nil, fmt.Errorf("failed to create commit message generator: %w", err)
	}
	return &LLMHook{Generator: generator, Config: cfg}, nil
}
