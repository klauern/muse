package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/klauern/pre-commit-llm/commit"
	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/llm"
	"github.com/klauern/pre-commit-llm/rag"
)

type PrepareCommitMsgHook interface {
	Run(commitMsgFile string, commitSource string, sha1 string) error
}

type LLMHook struct {
	Generator *commit.CommitMessageGenerator
}

func (h *LLMHook) Run(commitMsgFile string, commitSource string, sha1 string) error {
	// Always get the staged changes
	diff, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Get the commit style from the configuration
	commitStyle := h.Config.Hook.CommitStyle

	// Generate the commit message
	ctx := context.Background()
	message, err := h.Generator.Generate(ctx, diff, commitStyle)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Write the generated message to the commit message file
	if err := os.WriteFile(commitMsgFile, []byte(message), 0644); err != nil {
		return fmt.Errorf("failed to write commit message: %w", err)
	}

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

func NewHook(hookType string, cfg *config.Config) (PrepareCommitMsgHook, error) {
	switch hookType {
	case "llm":
		llmService, err := llm.NewLLMService(&cfg.LLM)
		if err != nil {
			return nil, fmt.Errorf("failed to create LLM service: %w", err)
		}

		ragService := &rag.GitRAGService{}
		generator := &commit.CommitMessageGenerator{
			LLMService: llmService,
			RAGService: ragService,
		}
		return &LLMHook{Generator: generator}, nil
	default:
		return &DefaultHook{}, nil
	}
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
