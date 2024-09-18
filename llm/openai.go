package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/your-project/templates"
)

type OpenAIProvider struct{}

func (p *OpenAIProvider) NewService(config map[string]interface{}) (LLMService, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not found in config")
	}
	model, ok := config["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("OpenAI model not found in config")
	}
	client := openai.NewClient(apiKey)
	return &OpenAIService{client: client, model: model}, nil
}

type OpenAIService struct {
	client *openai.Client
	model  string
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	commitTemplate := GetCommitTemplate(style)
	var promptBuffer bytes.Buffer
	err := commitTemplate.Template.Execute(&promptBuffer, struct {
		Diff    string
		Context string
		Schema  string
	}{
		Diff:    diff,
		Context: context,
		Schema:  commitTemplate.Schema.String(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: promptBuffer.String(),
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}

	// Parse the JSON response
	var commitMessage struct {
		Type    string `json:"type"`
		Scope   string `json:"scope"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Gitmoji string `json:"gitmoji,omitempty"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &commitMessage); err != nil {
		return "", fmt.Errorf("failed to parse commit message: %w", err)
	}

	// Format the commit message
	var formattedMessage string
	if style == GitmojisStyle {
		formattedMessage = fmt.Sprintf("%s %s", commitMessage.Gitmoji, commitMessage.Type)
	} else {
		formattedMessage = commitMessage.Type
	}
	if commitMessage.Scope != "" {
		formattedMessage += fmt.Sprintf("(%s)", commitMessage.Scope)
	}
	formattedMessage += fmt.Sprintf(": %s", commitMessage.Subject)
	if commitMessage.Body != "" {
		formattedMessage += fmt.Sprintf("\n\n%s", commitMessage.Body)
	}

	return formattedMessage, nil
}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}
