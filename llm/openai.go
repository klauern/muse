package llm

import (
	"context"
)

type OpenAIService struct {
	client *LLMClient
}

func NewOpenAIService(apiKey, modelName string) *OpenAIService {
	return &OpenAIService{
		client: NewLLMClient(apiKey, modelName),
	}
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	return s.client.GenerateCommitMessage(ctx, diff, context)
}
package llm

import (
	"context"
	"fmt"
)

type OpenAIProvider struct{}

func (p *OpenAIProvider) NewService(config map[string]interface{}) (LLMService, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok {
		return nil, fmt.Errorf("OpenAI API key not found in config")
	}
	model, ok := config["model"].(string)
	if !ok {
		return nil, fmt.Errorf("OpenAI model not found in config")
	}
	return &OpenAIService{apiKey: apiKey, model: model}, nil
}

type OpenAIService struct {
	apiKey string
	model  string
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	// TODO: Implement OpenAI API call
	return fmt.Sprintf("OpenAI generated commit message for diff: %s", diff), nil
}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}
