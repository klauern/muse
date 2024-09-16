package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
)

type AnthropicProvider struct {
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

func NewService(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = "claude-3-5-sonnet-20240620" // Default model if not specified
	}
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (s *AnthropicProvider) GenerateCommitMessage(ctx context.Context, diff, context string) (string, error) {
	prompt := fmt.Sprintf("Generate a concise commit message for the following diff:\n\n%s\n\nContext: %s", diff, context)

	req := Request{
		Model:     s.model,
		MaxTokens: 100, // Adjust as needed for commit message length
		Messages: []Message{
			{Role: "user", Content: prompt},
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
