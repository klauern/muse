package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

type AnthropicProvider struct{}

func (p *AnthropicProvider) NewService(config map[string]interface{}) (LLMService, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok {
		return nil, fmt.Errorf("Anthropic API key not found in config")
	}
	model, ok := config["model"].(string)
	if !ok {
		return nil, fmt.Errorf("Anthropic model not found in config")
	}
	client := anthropic.NewClient(apiKey)
	return &AnthropicService{client: client, model: model}, nil
}

type AnthropicService struct {
	client *anthropic.Client
	model  string
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	prompt := fmt.Sprintf("Given the following git diff and context, generate a concise and informative commit message:\n\nDiff:\n%s\n\nContext:\n%s", diff, context)
	
	resp, err := s.client.CreateCompletion(ctx, &anthropic.CompletionRequest{
		Model:     s.model,
		Prompt:    prompt,
		MaxTokens: 100,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return resp.Completion, nil
}

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
