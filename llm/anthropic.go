package llm

import (
	"context"
)

type AnthropicService struct {
	client *LLMClient
}

func NewAnthropicService(apiKey, modelName string) *AnthropicService {
	return &AnthropicService{
		client: NewLLMClient(apiKey, modelName),
	}
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	return s.client.GenerateCommitMessage(ctx, diff, context)
}
