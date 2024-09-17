package llm

import (
	"context"
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
)

type OllamaProvider struct{}

func (p *OllamaProvider) NewService(config *LLMConfig[OllamaConfig]) (LLMService, error) {
	if config.Config.Model == "" {
		return nil, fmt.Errorf("Ollama model not specified in config")
	}
	// TODO: Implement Ollama client initialization
	return &OllamaService{model: config.Config.Model}, nil
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
