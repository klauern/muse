package llm

import (
	"context"
	"fmt"

	"github.com/jmorganca/ollama/api"
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
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}
	return &OllamaService{client: client, model: model}, nil
}

type OllamaService struct {
	client *api.Client
	model  string
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	prompt := fmt.Sprintf("Given the following git diff and context, generate a concise and informative commit message:\n\nDiff:\n%s\n\nContext:\n%s", diff, context)
	
	resp, err := s.client.Generate(ctx, &api.GenerateRequest{
		Model:  s.model,
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return resp.Response, nil
}

func init() {
	RegisterProvider("ollama", &OllamaProvider{})
}
