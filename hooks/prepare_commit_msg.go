package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/llm"
	"github.com/klauern/pre-commit-llm/rag"
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
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Get the commit style from the configuration
	commitStyle := h.Config.HookConfig.CommitStyle

	// Generate the commit message
	ctx := context.Background()
	message, err := h.Generator.Generate(ctx, diff, commitStyle)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Check if dry run mode is enabled
	if h.Config.HookConfig.DryRun {
		fmt.Println("Dry run mode: Generated commit message:")
		fmt.Println(message)
		return nil
	}

	// Check if preview mode is enabled
	if h.Config.HookConfig.Preview {
		fmt.Println("Preview mode: Generated commit message:")
		fmt.Println(message)
		fmt.Print("Do you want to use this commit message? (y/n): ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if response != "y" && response != "Y" {
			return fmt.Errorf("user rejected the generated commit message")
		}
	}

	// Write the generated message to the commit message file
	if err := os.WriteFile(commitMsgFile, []byte(message), 0644); err != nil {
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
		return "", err
	}
	return string(output), nil
}

func NewHook(cfg *config.Config) (PrepareCommitMsgHook, error) {
	if cfg.HookConfig.Enabled {
		ragService := &rag.GitRAGService{}
		generator, err := llm.NewCommitMessageGenerator(cfg, ragService)
		if err != nil {
			return nil, fmt.Errorf("failed to create commit message generator: %w", err)
		}
		return &LLMHook{Generator: generator, Config: cfg}, nil
	}
	return &DefaultHook{}, nil
}

// DefaultHook is a no-op hook that doesn't modify the commit message
type DefaultHook struct{}

func (h *DefaultHook) Run(commitMsgFile string, commitSource string, sha1 string) error {
	// This hook doesn't modify the commit message, so we just return nil
	fmt.Println("DefaultHook: No modifications made to the commit message.")
	return nil
}
