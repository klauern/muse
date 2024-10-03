package llm

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/klauern/muse/api"
	"github.com/klauern/muse/templates"
)

type OpenAIProvider struct{}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}

type OpenAIService struct {
	client *api.Client
	model  string
}

func (p *OpenAIProvider) NewService(cfg map[string]interface{}) (LLMService, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	client := api.NewClient(apiKey)
	model := "gpt-3.5-turbo" // Default model, can be configurable

	return &OpenAIService{client: client, model: model}, nil
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff string, style CommitStyle) (string, error) {
	templateManager, err := templates.NewTemplateManager()
	if err != nil {
		return "", fmt.Errorf("failed to create template manager: %w", err)
	}
	style = GetCommitStyleFromString(style.String())

	var commitTemplate templates.CommitTemplate
	switch style {
	case ConventionalStyle:
		commitTemplate = templateManager.ConventionalCommit
	case GitmojisStyle:
		commitTemplate = templateManager.Gitmojis
	default:
		commitTemplate = templateManager.DefaultCommit
	}

	var promptBuffer bytes.Buffer
	err = commitTemplate.Template.Execute(&promptBuffer, map[string]interface{}{
		"Type": style.String(),
		"Diff": diff,
		"Format": func() []string {
			if prop, ok := commitTemplate.Schema.Definitions["ConventionalCommit"].Properties.Get("type"); ok {
				enumValues := make([]string, len(prop.Enum))
				for i, v := range prop.Enum {
					enumValues[i] = v.(string)
				}
				return enumValues
			}
			return nil
		}(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute commit template: %w", err)
	}

	req, err := s.client.NewRequest("POST", "/chat/completions", map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful assistant that generates commit messages.",
			},
			{
				"role":    "user",
				"content": promptBuffer.String(),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := s.client.Do(req, &response); err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no commit message generated")
	}

	return response.Choices[0].Message.Content, nil
}
