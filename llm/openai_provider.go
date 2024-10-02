package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	"github.com/klauern/muse/config"
	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct{}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}

type OpenAIService struct {
	client *instructor.Client
}

func (p *OpenAIProvider) NewService(cfg map[string]interface{}) (LLMService, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	model := "gpt-3.5-turbo"
	if modelCfg, ok := cfg["model"].(string); ok && modelCfg != "" {
		model = modelCfg
	}

	client, err := instructor.NewClient(apiKey, instructor.WithModel(model))
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	return &OpenAIService{client: client}, nil
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff string, style CommitStyle) (string, error) {
	template := GetCommitTemplate(style)

	prompt := fmt.Sprintf("Given the following git diff, generate a commit message in the %s style:\n\n%s", style, diff)

	var message GeneratedCommitMessage
	err := s.client.CreateCompletion(ctx, prompt, &message)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return message.String(), nil
}
