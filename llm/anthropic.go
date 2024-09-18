package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
)

type (
	AnthropicProvider struct{}

	AnthropicService struct {
		apiKey string
		model  string
	}

	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	Request struct {
		Model     string    `json:"model"`
		MaxTokens int       `json:"max_tokens"`
		Messages  []Message `json:"messages"`
		System    string    `json:"system"`
	}

	Response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	CommitMessage struct {
		Type    string      `json:"type"`
		Scope   string      `json:"scope"`
		Subject string      `json:"subject"`
		Body    interface{} `json:"body"`
	}
)

func (p *AnthropicProvider) NewService(config map[string]interface{}) (LLMService, error) {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API key not found in config or environment. Please set ANTHROPIC_API_KEY environment variable or provide it in the config")
		}
	}
	model, _ := config["model"].(string)
	if model == "" {
		model = "claude-3-5-sonnet-20240620" // Default model if not specified
	}
	return NewAnthropicService(apiKey, model), nil
}

func NewAnthropicService(apiKey, model string) *AnthropicService {
	return &AnthropicService{
		apiKey: apiKey,
		model:  model,
	}
}

func (s *AnthropicService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	if err := validateCommitStyle(style); err != nil {
		return "", err
	}

	systemPrompt, err := createSystemPrompt(diff, context, style)
	if err != nil {
		return "", err
	}

	req := createRequest(s.model, systemPrompt, diff, context)

	response, err := sendRequest(ctx, s.apiKey, req)
	if err != nil {
		return "", err
	}

	return formatCommitMessage(response)
}

func validateCommitStyle(style CommitStyle) error {
	if style < 0 || style.String() == "default" {
		return fmt.Errorf("invalid commit style: %v", style)
	}
	return nil
}

func createSystemPrompt(diff, context string, style CommitStyle) (string, error) {
	template := GetCommitTemplate(style)
	if template.Template == nil {
		return "", fmt.Errorf("invalid commit style: %v", style)
	}

	schemaJSON, err := json.Marshal(template.Schema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	var formatBuffer bytes.Buffer
	err = template.Template.Execute(&formatBuffer, struct {
		Type    string
		Diff    string
		Context string
		Format  string
		Details string
		Extra   string
		Schema  string
	}{
		Type:    template.Template.Name(),
		Diff:    diff,
		Context: context,
		Format:  "{{.Format}}",
		Details: "{{.Details}}",
		Extra:   "{{.Extra}}",
		Schema:  string(schemaJSON),
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return fmt.Sprintf("You are a Git commit message generator. Create a concise commit message based on the provided diff, following this format:\n%s\nEnsure the subject line (first line) is no longer than 72 characters. Complete the JSON structure below, filling in appropriate values for each field.", formatBuffer.String()), nil
}

func createRequest(model, systemPrompt, diff, context string) *Request {
	partialCompletion := `{
  "type": "`

	return &Request{
		Model:     model,
		MaxTokens: 200,
		System:    systemPrompt,
		Messages: []Message{
			{Role: "user", Content: fmt.Sprintf("Generate a commit message for this diff:\n\n%s\n\nAdditional context:\n%s", diff, context)},
			{Role: "assistant", Content: partialCompletion},
		},
	}
}

func sendRequest(ctx context.Context, apiKey string, req *Request) (*Response, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)
	httpReq.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 || response.Content[0].Type != "text" {
		return nil, fmt.Errorf("unexpected response format from API")
	}

	return &response, nil
}

func formatCommitMessage(response *Response) (string, error) {
	fullJSON := response.Content[0].Text

	// Attempt to parse the JSON as is
	var commitMessage CommitMessage
	err := json.Unmarshal([]byte(fullJSON), &commitMessage)
	if err != nil {
		// If parsing fails, try to extract the relevant information
		type_, scope, subject, body := extractCommitInfo(fullJSON)
		commitMessage = CommitMessage{
			Type:    type_,
			Scope:   scope,
			Subject: subject,
			Body:    body,
		}
	}

	// Ensure we have at least a type and subject
	if commitMessage.Type == "" {
		commitMessage.Type = "feat" // Default to "feat" if no type is provided
	}
	if commitMessage.Subject == "" {
		commitMessage.Subject = strings.SplitN(fullJSON, "\n", 2)[0] // Use the first line as subject if not found
	}

	var formattedMessage strings.Builder
	formattedMessage.WriteString(commitMessage.Type)
	if commitMessage.Scope != "" {
		formattedMessage.WriteString(fmt.Sprintf("(%s)", commitMessage.Scope))
	}
	formattedMessage.WriteString(": ")
	formattedMessage.WriteString(commitMessage.Subject)

	if commitMessage.Body != nil {
		formattedMessage.WriteString("\n\n")
		switch body := commitMessage.Body.(type) {
		case string:
			formattedMessage.WriteString(body)
		case []interface{}:
			for _, line := range body {
				if str, ok := line.(string); ok {
					formattedMessage.WriteString(str + "\n")
				}
			}
		}
	}

	return strings.TrimSpace(formattedMessage.String()), nil
}

func extractCommitInfo(jsonStr string) (type_ string, scope string, subject string, body string) {
	// If the string starts with a quote, it's likely the type
	if strings.HasPrefix(jsonStr, `"`) {
		parts := strings.SplitN(jsonStr, `"`, 3)
		if len(parts) > 1 {
			type_ = parts[1]
		}
	}

	scopeMatch := regexp.MustCompile(`"scope":\s*"([^"]+)"`).FindStringSubmatch(jsonStr)
	if len(scopeMatch) > 1 {
		scope = scopeMatch[1]
	}

	subjectMatch := regexp.MustCompile(`"subject":\s*"([^"]+)"`).FindStringSubmatch(jsonStr)
	if len(subjectMatch) > 1 {
		subject = subjectMatch[1]
	}

	bodyMatch := regexp.MustCompile(`"body":\s*"([^"]+)"`).FindStringSubmatch(jsonStr)
	if len(bodyMatch) > 1 {
		body = bodyMatch[1]
	}

	return
}

func init() {
	RegisterProvider("anthropic", &AnthropicProvider{})
}
