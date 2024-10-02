package llm

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"muse/config"
)

type OpenAIService struct {
	client *openai.Client
	model  string
}

func NewOpenAIService(cfg map[string]interface{}) (LLMService, error) {
	apiKey, ok := cfg["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is missing or invalid")
	}

	model, ok := cfg["model"].(string)
	if !ok || model == "" {
		model = "gpt-3.5-turbo" // Default model if not specified
	}

	apiBase, ok := cfg["api_base"].(string)
	if !ok || apiBase == "" {
		apiBase = openai.DefaultConfig("").BaseURL // Use default if not specified
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiBase

	client := openai.NewClientWithConfig(config)

	return &OpenAIService{
		client: client,
		model:  model,
	}, nil
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff string, style CommitStyle) (string, error) {
	prompt := generatePrompt(diff, style)

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func generatePrompt(diff string, style CommitStyle) string {
	basePrompt := fmt.Sprintf("Given the following git diff, generate a commit message:\n\n%s\n\n", diff)

	switch style {
	case ConventionalCommit:
		return basePrompt + "Please format the commit message following the Conventional Commits specification."
	case DetailedCommit:
		return basePrompt + "Please provide a detailed commit message with a summary and bullet points for changes."
	default:
		return basePrompt + "Please provide a concise and informative commit message."
	}
}
