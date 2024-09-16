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

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	template := GetCommitTemplate(style)
	var promptBuffer bytes.Buffer
	err := template.Execute(&promptBuffer, struct {
		Diff    string
		Context string
	}{
		Diff:    diff,
		Context: context,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// TODO: Implement OpenAI API call using the generated prompt
	return fmt.Sprintf("OpenAI generated commit message for diff: %s", diff), nil
}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}
