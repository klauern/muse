package llm

import (
	"context"
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
)

// LLMService defines the interface for LLM providers
type LLMService interface {
	GenerateCommitMessage(ctx context.Context, diff, context string) (string, error)
}

// NewLLMService creates a new LLMService based on the provided configuration
func NewLLMService(cfg *config.LLMConfig) (LLMService, error) {
	switch cfg.Provider {
	case "openai":
		return NewOpenAIService(cfg.OpenAIAPIKey, cfg.OpenAIModel), nil
	case "anthropic":
		return NewAnthropicService(cfg.AnthropicAPIKey, cfg.AnthropicModel), nil
	case "ollama":
		return NewOllamaService(cfg.OllamaEndpoint, cfg.OllamaModel), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}
