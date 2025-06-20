package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/klauern/muse/templates"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct{}

func init() {
	RegisterProvider("openai", &OpenAIProvider{})
}

type OpenAIService struct {
	client  *openai.Client
	model   string
	apiKey  string
	apiBase string
}

func (p *OpenAIProvider) NewService(cfg map[string]any) (LLMService, error) {
	// Try to get API key from config, then environment variables
	apiKey, _ := cfg["api_key"].(string)
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			slog.Error("OpenAI API key not set in config or OPENAI_API_KEY environment variable")
			return nil, fmt.Errorf("openai api key not set")
		}
		slog.Debug("Using OPENAI_API_KEY environment variable")
	}

	options := []option.RequestOption{option.WithAPIKey(apiKey)}

	// Try to get API base from config, then environment variables
	apiBase, _ := cfg["api_base"].(string)
	if apiBase == "" {
		apiBase = os.Getenv("OPENAI_API_BASE")
		if apiBase == "" {
			apiBase = "https://api.openai.com/v1"
		}
		slog.Debug("Using API base from environment or default", "api_base", apiBase)
	}
	options = append(options, option.WithBaseURL(apiBase))

	client := openai.NewClient(options...)

	// Get model from config with fallback
	model, _ := cfg["model"].(string)
	if model == "" {
		slog.Warn("No model specified, using default gpt-4o")
		model = "gpt-4o"
	}
	slog.Debug("Using model", "model", model)

	return &OpenAIService{
		client:  client,
		model:   model,
		apiKey:  apiKey,
		apiBase: apiBase,
	}, nil
}

func (s *OpenAIService) GenerateCommitMessage(ctx context.Context, diff string, style templates.CommitStyle) (string, error) {
	templateManager := templates.NewTemplateManager(diff, style)

	commitTemplate, err := templateManager.CompileTemplate(style)
	slog.Debug("Commit template", "template", commitTemplate.Template.Root.String())
	if err != nil {
		slog.Error("Failed to compile template", "error", err)
		return "", fmt.Errorf("failed to execute commit template: %w", err)
	}

	// Use raw HTTP for gpt-4.1 to handle API gateway content-type issues
	if s.model == "gpt-4.1" {
		slog.Debug("Using raw HTTP client for gpt-4.1 due to API gateway compatibility")
		return s.generateWithRawHTTP(ctx, commitTemplate)
	}

	// Try structured outputs first for compatible models, but fall back on error
	if s.supportsStructuredOutputs() {
		result, err := s.generateWithStructuredOutputs(ctx, commitTemplate)
		if err == nil {
			return result, nil
		}
		slog.Warn("Structured outputs failed, falling back to regular completion", "error", err)
	}

	// Fallback to regular chat completion
	return s.generateWithRegularCompletion(ctx, commitTemplate)
}

// supportsStructuredOutputs checks if the current model supports structured outputs
var structuredOutputModels = map[string]bool{
	"gpt-4o-2024-08-06":      true,
	"gpt-4o-mini-2024-07-18": true,
	"gpt-4o-2024-11-20":      true,
	"gpt-4o-mini":            true,
	"gpt-4o":                 true,
	// Disabled gpt-4.1 structured outputs due to API gateway compatibility issues
	// "gpt-4.1":                true,
}

func (s *OpenAIService) supportsStructuredOutputs() bool {
	return structuredOutputModels[s.model]
}

// generateWithStructuredOutputs uses OpenAI's structured outputs
func (s *OpenAIService) generateWithStructuredOutputs(ctx context.Context, commitTemplate templates.CommitTemplate) (string, error) {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("CommitDiffInstructions"),
		Description: openai.F("Commit instructions for the diff"),
		Strict:      openai.Bool(true),
		Schema:      openai.F(templates.CommitStyleTemplateSchema),
	}

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
		Model: openai.F(s.model),
	})
	if err != nil {
		slog.Debug("Structured outputs error details", "error", err)
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	conventionalCommit := templates.ConventionalCommit{}
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &conventionalCommit)
	if err != nil {
		slog.Error("Failed to unmarshal structured chat completion", "error", err)
		return "", fmt.Errorf("failed to unmarshal chat completion: %w", err)
	}
	return conventionalCommit.String(), nil
}

// generateWithRegularCompletion uses regular chat completion without JSON mode for gateway compatibility
func (s *OpenAIService) generateWithRegularCompletion(ctx context.Context, commitTemplate templates.CommitTemplate) (string, error) {
	// Create a simpler prompt that asks for a direct commit message
	simplePrompt := fmt.Sprintf(`Analyze the following git diff and generate a conventional commit message:

%s

Generate a single-line conventional commit message in the format:
<type>(<scope>): <description>

Where:
- type: feat, fix, docs, style, refactor, test, chore, build, ci, perf, or revert
- scope: optional, the part of the codebase affected (keep it short)
- description: brief description of the change in present tense

Examples:
- feat(auth): add OAuth2 login support
- fix(api): resolve timeout issues in user service
- docs(readme): update installation instructions

Respond with ONLY the commit message, no additional text or formatting.`,
		commitTemplate.Template.Root.String())

	chat, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(simplePrompt),
		}),
		Model: openai.F(s.model),
		// Remove JSON mode to avoid content-type issues with your gateway
	})
	if err != nil {
		slog.Debug("Regular completion error details", "error", err)
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Return the raw response as the commit message
	commitMessage := strings.TrimSpace(chat.Choices[0].Message.Content)
	slog.Debug("Generated commit message", "message", commitMessage)
	return commitMessage, nil
}

// generateWithRawHTTP makes a direct HTTP request to handle API gateway content-type issues
func (s *OpenAIService) generateWithRawHTTP(ctx context.Context, commitTemplate templates.CommitTemplate) (string, error) {
	// Create a simple prompt for conventional commit messages
	simplePrompt := fmt.Sprintf(`Analyze the following git diff and generate a conventional commit message:

%s

Generate a single-line conventional commit message in the format:
<type>(<scope>): <description>

Where:
- type: feat, fix, docs, style, refactor, test, chore, build, ci, perf, or revert  
- scope: optional, the part of the codebase affected (keep it short)
- description: brief description of the change in present tense

Examples:
- feat(auth): add OAuth2 login support
- fix(api): resolve timeout issues in user service
- docs(readme): update installation instructions

Respond with ONLY the commit message, no additional text or formatting.`,
		commitTemplate.Template.Root.String())

	// Prepare the request body
	requestBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": simplePrompt,
			},
		},
		"max_tokens":  100,
		"temperature": 0.7,
	}

	bodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	url := strings.TrimSuffix(s.apiBase, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("User-Agent", "muse-commit-generator/1.0")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("Raw HTTP response", "status", resp.Status, "content-type", resp.Header.Get("Content-Type"), "body", string(bodyBytes))

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Try to parse as JSON first
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err == nil {
		// Successfully parsed as JSON - extract the message
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						commitMessage := strings.TrimSpace(content)
						slog.Debug("Generated commit message via raw HTTP", "message", commitMessage)
						return commitMessage, nil
					}
				}
			}
		}
		return "", fmt.Errorf("unable to extract message content from JSON response")
	}

	// If JSON parsing failed, treat the response as plain text
	responseText := strings.TrimSpace(string(bodyBytes))
	slog.Debug("Treating response as plain text", "response", responseText)

	// Extract commit message from plain text response
	lines := strings.Split(responseText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "{") && !strings.HasPrefix(line, "[") {
			// Found a non-empty, non-JSON line - likely our commit message
			slog.Debug("Generated commit message via raw HTTP (plain text)", "message", line)
			return line, nil
		}
	}

	return "", fmt.Errorf("unable to extract commit message from response: %s", responseText)
}
