package llm

import (
	"context"
	"fmt"
)

type OllamaProvider struct{}

func (p *OllamaProvider) NewService(config *config.LLMConfig) (LLMService, error) {
	// TODO: Implement Ollama client initialization
	return nil, fmt.Errorf("Ollama provider not implemented")
}

type OllamaService struct {
	// TODO: Add necessary fields
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Ollama API call
	return "", fmt.Errorf("Ollama service not implemented")
}

func init() {
	RegisterProvider("ollama", &OllamaProvider{})
}
