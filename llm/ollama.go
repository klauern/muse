package llm

import (
	"context"
	"fmt"
)

type OllamaService struct {
	Endpoint string
	ModelName string
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Ollama API call
	return fmt.Sprintf("Ollama generated commit message for diff: %s", diff), nil
}
