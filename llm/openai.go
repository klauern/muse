package llm

import (
	"context"
	"fmt"
)

type OpenAIService struct {
	APIKey   string
	ModelName string
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement OpenAI API call
	return fmt.Sprintf("OpenAI generated commit message for diff: %s", diff), nil
}
