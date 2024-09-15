package llm

import (
	"context"
)

type AnthropicService struct {
	apiKey    string
	modelName string
}

func NewAnthropicService(apiKey, modelName string) *AnthropicService {
	return &AnthropicService{
		apiKey:    apiKey,
		modelName: modelName,
	}
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	return s.client.GenerateCommitMessage(ctx, diff, context)
}
