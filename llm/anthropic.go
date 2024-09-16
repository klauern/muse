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
	System    string    `json:"system"`
}

type Response struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
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
		Scope   string
		Subject string
		Body    string
		Footer  string
		Extra   map[string]string
	}{
		Diff:    diff,
		Context: context,
		Type:    "feat", // Default to "feat" for now, you might want to determine this dynamically
		Format:  style.String(),
		Details: diff, // Use the diff as details for now
		Scope:   "",   // You might want to determine this based on the diff
		Subject: "",   // This will be filled by the LLM
		Body:    "",   // This will be filled by the LLM
		Footer:  "",   // This will be filled by the LLM
		Extra:   make(map[string]string), // Initialize an empty map for any extra fields
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	req := Request{
		Model:     s.model,
		MaxTokens: 300, // Reduced to encourage more concise responses
		System:    "You are a Git commit message generator. Create a concise, conventional commit message based on the provided diff. The message should have a brief subject line (type(scope): description) followed by a blank line and a short bullet list of key changes. Exclude file names, line numbers, and diff syntax. Focus only on the most important changes.",
		Messages: []Message{
			{Role: "user", Content: "Generate a commit message for this diff:\n\n" + diff},
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

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Extract the commit message from the fenced code block
	fullResponse := response.Content[0].Text
	commitMessage := extractCommitMessage(fullResponse)

	return commitMessage, nil
}

func extractCommitMessage(response string) string {
	// Remove any markdown code block markers
	response = strings.ReplaceAll(response, "```", "")

	// Trim any leading or trailing whitespace
	response = strings.TrimSpace(response)

	// Split the response into lines
	lines := strings.Split(response, "\n")

	// Ensure we have at least a subject line
	if len(lines) == 0 {
		return ""
	}

	// Keep the subject line and up to 5 bullet points
	result := []string{lines[0]}
	bulletPoints := 0
	for i := 1; i < len(lines) && bulletPoints < 5; i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			result = append(result, line)
			bulletPoints++
		}
	}

	// Join the lines back together
	return strings.Join(result, "\n")
}

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
