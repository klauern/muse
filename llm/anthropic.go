package llm

import (
	"context"
	"fmt"
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
	return &AnthropicService{apiKey: apiKey, model: model}, nil
}

type AnthropicService struct {
	apiKey string
	model  string
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Anthropic API call
	return fmt.Sprintf("Anthropic generated commit message for diff: %s", diff), nil
}

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
