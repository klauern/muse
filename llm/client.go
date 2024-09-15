package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type LLMClient struct {
	client    *openai.Client
	modelName string
}

func NewLLMClient(apiKey, modelName string) *LLMClient {
	return &LLMClient{
		client:    openai.NewClient(apiKey),
		modelName: modelName,
	}
}

func (c *LLMClient) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	prompt := fmt.Sprintf("Given the following git diff and context, generate a concise and informative commit message:\n\nDiff:\n%s\n\nContext:\n%s", diff, context)

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.modelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("error generating commit message: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}
