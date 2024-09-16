package llm

import (
	"context"
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
)

type LLMClient struct {
	service LLMService
}

func NewLLMClient(cfg *config.LLMConfig) (*LLMClient, error) {
	service, err := NewLLMService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return &LLMClient{
		service: service,
	}, nil
}

func (c *LLMClient) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	return c.service.GenerateCommitMessage(ctx, diff, context, style)
}
