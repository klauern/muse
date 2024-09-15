package llm

import (
	"context"
	"fmt"
)

type AnthropicProvider struct{}

func (p *AnthropicProvider) NewService(config map[string]interface{}) (LLMService, error) {
	// TODO: Implement Anthropic client initialization
	return nil, fmt.Errorf("Anthropic provider not implemented")
}

type AnthropicService struct {
	// TODO: Add necessary fields
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement Anthropic API call
	return "", fmt.Errorf("Anthropic service not implemented")
}

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
