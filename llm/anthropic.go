package llm

import (
	"context"
	"fmt"
)

type AnthropicService struct {
	APIKey   string
	ModelName string
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Anthropic API call
	return fmt.Sprintf("Anthropic generated commit message for diff: %s", diff), nil
}
