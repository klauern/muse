package llm

import (
	"context"
	"fmt"
)

type OllamaProvider struct{}

func (p *OllamaProvider) NewService(config map[string]interface{}) (LLMService, error) {
	endpoint, ok := config["endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("Ollama endpoint not found in config")
	}
	model, ok := config["model"].(string)
	if !ok {
		return nil, fmt.Errorf("Ollama model not found in config")
	}
	return &OllamaService{endpoint: endpoint, model: model}, nil
}

type OllamaService struct {
	endpoint string
	model    string
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Ollama API call
	return fmt.Sprintf("Ollama generated commit message for diff: %s", diff), nil
}

func init() {
	RegisterProvider("ollama", &OllamaProvider{})
}
