package llm

import (
	"context"
	"fmt"
)

type OllamaProvider struct{}

func (p *OllamaProvider) NewService(config map[string]interface{}) (LLMService, error) {
	model, ok := config["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("Ollama model not specified in config")
	}
	// TODO: Implement Ollama client initialization
	return &OllamaService{model: model}, nil
}

type OllamaService struct {
	model string
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	// TODO: Implement Ollama API call
	return "", fmt.Errorf("Ollama service not implemented")
}

func init() {
	RegisterProvider("ollama", &OllamaProvider{})
}
