package llm

import (
	"context"
	"fmt"
)

type OllamaService struct {
	Endpoint  string
	ModelName string
}

func NewOllamaService(endpoint, modelName string) *OllamaService {
	return &OllamaService{
		Endpoint:  endpoint,
		ModelName: modelName,
	}
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Ollama API call using a custom HTTP client
	return fmt.Sprintf("Ollama generated commit message for diff: %s", diff), nil
}
