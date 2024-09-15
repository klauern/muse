package llm

import (
	"context"
)

type OpenAIService struct {
	client *LLMClient
}

func NewOpenAIService(apiKey, modelName string) *OpenAIService {
	return &OpenAIService{
		client: NewLLMClient(apiKey, modelName),
	}
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	return s.client.GenerateCommitMessage(ctx, diff, context)
}
