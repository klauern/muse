package llm

import "context"

// LLMService defines the interface for LLM providers
type LLMService interface {
	GenerateCommitMessage(ctx context.Context, diff, context string) (string, error)
}
