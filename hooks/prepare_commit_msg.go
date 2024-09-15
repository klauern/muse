package pre_commit_llm

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/klauern/pre-commit-llm/commit"
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
    var diff string
    var err error

    switch commitSource {
    case "message":
        // Normal commit
        diff, err = getGitDiff("--cached")
    case "commit":
        // Amending a commit
        diff, err = getGitDiff("HEAD^")
    case "merge":
        // Merge commit
        diff, err = getMergeDiff()
    default:
        // Default to cached diff
        diff, err = getGitDiff("--cached")
    }

    if err != nil {
        return fmt.Errorf("failed to get git diff: %w", err)
    }

    // Generate the commit message
    ctx := context.Background()
    message, err := h.Generator.Generate(ctx, diff)
    if err != nil {
        return fmt.Errorf("failed to generate commit message: %w", err)
    }

    // Write the generated message to the commit message file
    if err := os.WriteFile(commitMsgFile, []byte(message), 0644); err != nil {
        return fmt.Errorf("failed to write commit message: %w", err)
    }

    return nil
}

func getGitDiff(target string) (string, error) {
    cmd := exec.Command("git", "diff", target)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}

func getMergeDiff() (string, error) {
    cmd := exec.Command("git", "diff", "HEAD^1", "HEAD^2")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}

func NewHook(hookType string, config *Config) PrepareCommitMsgHook {
	switch hookType {
	case "llm":
		var llmService llm.LLMService
		switch config.LLM.Provider {
		case "openai":
			llmService = llm.NewOpenAIService(config.LLM.OpenAIAPIKey, config.LLM.OpenAIModel)
		case "anthropic":
			llmService = llm.NewAnthropicService(config.LLM.AnthropicAPIKey, config.LLM.AnthropicModel)
		case "ollama":
			llmService = llm.NewOllamaService(config.LLM.OllamaEndpoint, config.LLM.OllamaModel)
		default:
			llmService = llm.NewOpenAIService(config.LLM.OpenAIAPIKey, config.LLM.OpenAIModel)
		}

		ragService := &rag.GitRAGService{}
		generator := &commit.CommitMessageGenerator{
			LLMService: llmService,
			RAGService: ragService,
		}
		return &LLMHook{Generator: generator}
	default:
		return &DefaultHook{}
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
