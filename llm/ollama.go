package llm

import (
	"context"
	"fmt"
)

type OllamaService struct {
	endpoint  string
	modelName string
}

func NewOllamaService(endpoint, modelName string) *OllamaService {
	return &OllamaService{
		endpoint:  endpoint,
		modelName: modelName,
	}
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Ollama API call using a custom HTTP client
	return fmt.Sprintf("Ollama generated commit message for diff: %s", diff), nil
}
