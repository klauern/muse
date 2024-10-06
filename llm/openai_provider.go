package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/klauern/muse/templates"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct{}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}

type OpenAIService struct {
	client *openai.Client
	model  string
}

func (p *OpenAIProvider) NewService(cfg map[string]any) (LLMService, error) {
	apiKey, ok := cfg["api_key"].(string)
	if !ok {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		slog.Error("OpenAI API key not set")
		return nil, fmt.Errorf("openai api key not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	model := "gpt-4o" // Default model, can be configurable

	return &OpenAIService{client: client, model: model}, nil
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff string, style templates.CommitStyle) (string, error) {
	templateManager := templates.NewTemplateManager(diff, style)

	commitTemplate, err := templateManager.CompileTemplate(style)
	if err != nil {
		slog.Error("Failed to compile template", "error", err)
		return "", fmt.Errorf("failed to execute commit template: %w", err)
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("CommitDiffInstructions"),
		Description: openai.F("Commit instructions for the diff"),
		Strict:      openai.Bool(true),
		Schema:      openai.F(templates.CommitStyleTemplateSchema),
	}

	// Query the Chat Completions API
	chat, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(commitTemplate.Template.Root.String()),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		// Only certain models can perform structured outputs
		Model: openai.F(openai.ChatModelGPT4o2024_08_06),
	})
	if err != nil {
		slog.Error("Failed to create chat completion", "error", err)
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	// The model responds with a JSON string, so parse it into a struct
	conventionalCommit := templates.ConventionalCommit{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &conventionalCommit)
	if err != nil {
		slog.Error("Failed to unmarshal chat completion", "error", err)
		return "", fmt.Errorf("failed to unmarshal chat completion: %w", err)
	}
	return conventionalCommit.String(), nil
}
