package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
)

type AnthropicProvider struct{}

func (p *AnthropicProvider) NewService(config map[string]interface{}) (LLMService, error) {
	apiKey := config.Config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API key not found in config or environment. Please set ANTHROPIC_API_KEY environment variable or provide it in the config")
		}
	}
	model := config.Config.Model
	if model == "" {
		model = "claude-3-sonnet-20240229" // Default model if not specified
	}
	return NewAnthropicService(apiKey, model), nil
}

type AnthropicService struct {
	apiKey string
	model  string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system"`
}

type Response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type CommitMessage struct {
	Type    string      `json:"type"`
	Scope   string      `json:"scope"`
	Subject string      `json:"subject"`
	Body    interface{} `json:"body"`
}

func NewAnthropicService(apiKey, model string) *AnthropicService {
	return &AnthropicService{
		apiKey: apiKey,
		model:  model,
	}
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	template := GetCommitTemplate(style)
	if template == nil {
		return "", fmt.Errorf("invalid commit style: %v", style)
	}

	// Early return if the style is invalid
	if style < 0 || style.String() == "default" {
		return "", fmt.Errorf("invalid commit style: %v", style)
	}

	var formatBuffer bytes.Buffer
	err := template.Execute(&formatBuffer, struct {
		Type    string
		Diff    string
		Context string
		Format  string
		Details string
		Extra   string
	}{
		Type:    template.Name(),
		Diff:    diff,
		Context: context,
		Format:  "{{.Format}}",
		Details: "{{.Details}}",
		Extra:   "{{.Extra}}",
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	systemPrompt := fmt.Sprintf("You are a Git commit message generator. Create a concise commit message based on the provided diff, following this format:\n%s\nEnsure the subject line (first line) is no longer than 72 characters. Complete the JSON structure below, filling in appropriate values for each field.", formatBuffer.String())

	partialCompletion := `{
  "type": "`

	req := Request{
		Model:     s.model,
		MaxTokens: 200,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: fmt.Sprintf("Generate a commit message for this diff:\n\n%s\n\nAdditional context:\n%s", diff, context)},
			{Role: "assistant", Content: partialCompletion},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("x-api-key", s.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)
	httpReq.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	if len(response.Content) == 0 || response.Content[0].Type != "text" {
		return "", fmt.Errorf("unexpected response format from API")
	}

	// Combine the partial completion with the response to get the full JSON
	fullJSON := partialCompletion + response.Content[0].Text

	// Attempt to parse the JSON into a CommitMessage struct
	var commitMessage struct {
		Type    string      `json:"type"`
		Scope   string      `json:"scope"`
		Subject string      `json:"subject"`
		Body    interface{} `json:"body"`
	}
	err = json.Unmarshal([]byte(fullJSON), &commitMessage)
	if err != nil {
		// If parsing fails, return the raw response for debugging
		return "", fmt.Errorf("failed to parse commit message JSON: %w\nRaw response: %s", err, fullJSON)
	}

	// Format the commit message
	var formattedMessage strings.Builder
	formattedMessage.WriteString(fmt.Sprintf("%s", commitMessage.Type))
	if commitMessage.Scope != "" {
		formattedMessage.WriteString(fmt.Sprintf("(%s)", commitMessage.Scope))
	}
	formattedMessage.WriteString(fmt.Sprintf(": %s\n\n", commitMessage.Subject))

	// Handle body based on its type
	switch body := commitMessage.Body.(type) {
	case string:
		formattedMessage.WriteString(fmt.Sprintf("%s\n", body))
	case []interface{}:
		for _, line := range body {
			if str, ok := line.(string); ok {
				formattedMessage.WriteString(fmt.Sprintf("%s\n", str))
			}
		}
	}

	return strings.TrimSpace(formattedMessage.String()), nil
}

// The extractCommitMessage function is no longer needed as we're parsing JSON directly

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
