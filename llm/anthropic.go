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

	"github.com/klauern/pre-commit-llm/config"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
)

type AnthropicProvider struct{}

func (p *AnthropicProvider) NewService(config *config.LLMConfig) (LLMService, error) {
	apiKey, ok := config.Extra["api_key"].(string)
	if !ok || apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API key not found in config or environment. Please set ANTHROPIC_API_KEY environment variable or provide it in the config")
		}
	}
	model := config.Model
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
}

type Response struct {
	Content string `json:"content"`
}

func NewAnthropicService(apiKey, model string) *AnthropicService {
	return &AnthropicService{
		apiKey: apiKey,
		model:  model,
	}
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	template := GetCommitTemplate(style)
	var promptBuffer bytes.Buffer
	err := template.Execute(&promptBuffer, struct {
		Diff    string
		Context string
		Type    string
		Format  string
		Details string
	}{
		Diff:    diff,
		Context: context,
		Type:    "feat", // Default to "feat" for now, you might want to determine this dynamically
		Format:  style.String(),
		Details: diff, // Use the diff as details for now
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	req := Request{
		Model:     s.model,
		MaxTokens: 300, // Increased to allow for a more detailed commit message
		Messages: []Message{
			{Role: "user", Content: promptBuffer.String()},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBody)))
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

	return strings.TrimSpace(response.Content), nil
}

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
