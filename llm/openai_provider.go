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
	"regexp"
	"strings"
	"time"

	"github.com/klauern/muse/internal/security"
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
	}

	// Validate the API key
	if err := security.ValidateCredential(apiKey); err != nil {
		slog.Warn("API key validation warning", "issue", err.Error(), "masked_key", security.MaskCredential(apiKey))
		// Continue anyway as it might still work, but warn the user
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
	if err != nil {
		slog.Error("Failed to compile template", "error", err)
		return "", fmt.Errorf("failed to execute commit template: %w", err)
	}

	// Execute template with data to create the final prompt
	prompt, err := s.executeTemplate(commitTemplate, templateManager)
	if err != nil {
		slog.Error("Failed to execute template", "error", err)
		return "", fmt.Errorf("failed to execute commit template: %w", err)
	}
	slog.Debug("Generated prompt from template", "length", len(prompt))

	// Use raw HTTP for gpt-4.1 to handle API gateway content-type issues
	if s.model == "gpt-4.1" {
		slog.Debug("Using raw HTTP client for gpt-4.1 due to API gateway compatibility")
		return s.generateWithRawHTTP(ctx, commitTemplate, templateManager)
	}

	// Try structured outputs first for compatible models, but fall back on error
	if s.supportsStructuredOutputs() {
		result, err := s.generateWithStructuredOutputs(ctx, commitTemplate, templateManager)
		if err == nil {
			return result, nil
		}
		// Check if this is a content-type issue that requires raw HTTP
		if s.isContentTypeError(err) {
			slog.Warn("Structured outputs failed due to content-type issue, falling back to raw HTTP", "error", err)
			return s.generateWithRawHTTP(ctx, commitTemplate, templateManager)
		}
		slog.Warn("Structured outputs failed, falling back to regular completion", "error", err)
	}

	// Fallback to regular chat completion
	result, err := s.generateWithRegularCompletion(ctx, commitTemplate, templateManager)
	if err != nil {
		// Check if this is a content-type issue that requires raw HTTP
		if s.isContentTypeError(err) {
			slog.Warn("Regular completion failed due to content-type issue, falling back to raw HTTP", "error", err)
			return s.generateWithRawHTTP(ctx, commitTemplate, templateManager)
		}
		return "", err
	}
	return result, nil
}

// executeTemplate executes the template with data to generate the final prompt
func (s *OpenAIService) executeTemplate(commitTemplate templates.CommitTemplate, templateManager *templates.TemplateManager) (string, error) {
	data := templateManager.GetTemplateData()

	var buf strings.Builder
	err := commitTemplate.Template.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// structuredOutputModels defines which models support structured outputs
var structuredOutputModels = map[string]bool{
	"gpt-4o-2024-08-06":      true,
	"gpt-4o-mini-2024-07-18": true,
	"gpt-4o-2024-11-20":      true,
	"gpt-4o-mini":            true,
	"gpt-4o":                 true,
	// Disabled gpt-4.1 structured outputs due to API gateway compatibility issues
	// "gpt-4.1":                true,
}

// supportsStructuredOutputs checks if the current model supports structured outputs
func (s *OpenAIService) supportsStructuredOutputs() bool {
	return structuredOutputModels[s.model]
}

// isContentTypeError checks if the error is related to content-type issues
func (s *OpenAIService) isContentTypeError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "content-type") ||
		strings.Contains(errStr, "application/json") ||
		strings.Contains(errStr, "expected destination type")
}

// generateWithStructuredOutputs uses OpenAI's structured outputs
func (s *OpenAIService) generateWithStructuredOutputs(ctx context.Context, commitTemplate templates.CommitTemplate, templateManager *templates.TemplateManager) (string, error) {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("CommitDiffInstructions"),
		Description: openai.F("Commit instructions for the diff"),
		Strict:      openai.Bool(true),
		Schema:      openai.F(interface{}(commitTemplate.Schema)),
	}

	// Execute template with data first
	prompt, err := s.executeTemplate(commitTemplate, templateManager)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	chat, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
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
func (s *OpenAIService) generateWithRegularCompletion(ctx context.Context, commitTemplate templates.CommitTemplate, templateManager *templates.TemplateManager) (string, error) {
	// Execute template with data to create the final prompt
	prompt, err := s.executeTemplate(commitTemplate, templateManager)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	chat, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
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
func (s *OpenAIService) generateWithRawHTTP(ctx context.Context, commitTemplate templates.CommitTemplate, templateManager *templates.TemplateManager) (string, error) {
	// Execute template with data to create the final prompt
	prompt, err := s.executeTemplate(commitTemplate, templateManager)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Prepare the request body
	requestBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
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

	// Make the request with timeout
	client := &http.Client{
		Timeout: 60 * time.Second, // 60 second timeout
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Failed to close response body", "error", err)
		}
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("Raw HTTP response", "status", resp.Status, "content-type", resp.Header.Get("Content-Type"), "body_length", len(bodyBytes))
	slog.Debug("Raw HTTP response body", "body", string(bodyBytes))

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		slog.Error("API request failed", "status_code", resp.StatusCode, "response", string(bodyBytes))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Check for empty response
	if len(bodyBytes) == 0 {
		slog.Error("Received empty response from API")
		return "", fmt.Errorf("received empty response from API")
	}

	// Try to parse as JSON first
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err == nil {
		// Successfully parsed as JSON - extract the message
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						slog.Debug("Extracted content from API response", "content", content)
						commitMessage := s.extractCommitMessage(content)
						slog.Debug("Generated commit message via raw HTTP", "message", commitMessage, "length", len(commitMessage))
						if commitMessage == "" {
							slog.Error("Extracted commit message is empty", "original_content", content)
							return "", fmt.Errorf("extracted commit message is empty from content: %s", content)
						}
						return commitMessage, nil
					}
				}
			}
		}
		return "", fmt.Errorf("unable to extract message content from JSON response")
	}

	// If JSON parsing failed, treat the response as plain text
	responseText := strings.TrimSpace(string(bodyBytes))
	slog.Debug("Treating response as plain text", "response", responseText, "length", len(responseText))

	commitMessage := s.extractCommitMessage(responseText)
	slog.Debug("Extracted commit message from plain text", "message", commitMessage, "length", len(commitMessage))
	if commitMessage != "" {
		slog.Debug("Generated commit message via raw HTTP (plain text)", "message", commitMessage)
		return commitMessage, nil
	}

	slog.Error("Unable to extract commit message", "response_text", responseText)
	return "", fmt.Errorf("unable to extract commit message from response: %s", responseText)
}

// extractCommitMessage extracts a clean commit message from various response formats
func (s *OpenAIService) extractCommitMessage(content string) string {
	content = strings.TrimSpace(content)

	// Try to parse as JSON first - handle structured commit format
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonData); err == nil {
		slog.Debug("Successfully parsed JSON response", "keys", getJSONKeys(jsonData))

		// Handle structured commit format: {"type": "feat", "scope": "api", "subject": "...", "body": "..."}
		if commitType, hasType := jsonData["type"].(string); hasType {
			if subject, hasSubject := jsonData["subject"].(string); hasSubject {
				message := commitType
				if scope, hasScope := jsonData["scope"].(string); hasScope && scope != "" {
					message += "(" + scope + ")"
				}
				message += ": " + subject

				if body, hasBody := jsonData["body"].(string); hasBody && body != "" {
					message += "\n\n" + body
				}

				if footer, hasFooter := jsonData["footer"].(string); hasFooter && footer != "" {
					message += "\n\n" + footer
				}

				slog.Debug("Built commit message from structured format", "message", message)
				return strings.TrimSpace(message)
			}
		}

		// Handle legacy commit_message format
		if msg, ok := jsonData["commit_message"].(string); ok {
			return strings.TrimSpace(msg)
		}
	} else {
		slog.Debug("Failed to parse as complete JSON, checking for truncated response", "error", err)

		// Try to handle truncated JSON by attempting to fix common issues
		if fixedContent := s.tryFixTruncatedJSON(content); fixedContent != content {
			if err := json.Unmarshal([]byte(fixedContent), &jsonData); err == nil {
				slog.Debug("Successfully parsed fixed JSON response", "keys", getJSONKeys(jsonData))

				// Handle structured commit format from fixed JSON
				if commitType, hasType := jsonData["type"].(string); hasType {
					if subject, hasSubject := jsonData["subject"].(string); hasSubject {
						message := commitType
						if scope, hasScope := jsonData["scope"].(string); hasScope && scope != "" {
							message += "(" + scope + ")"
						}
						message += ": " + subject

						if body, hasBody := jsonData["body"].(string); hasBody && body != "" {
							message += "\n\n" + body
						}

						slog.Debug("Built commit message from fixed truncated JSON", "message", message)
						return strings.TrimSpace(message)
					}
				}
			}
		}
	}

	// Try to extract from JSON-formatted responses in markdown code blocks like:
	// ```json\n{\n  "type": "feat", "subject": "add new feature"\n}\n```
	if strings.Contains(content, "```json") || (strings.Contains(content, "type") && strings.Contains(content, "subject")) {
		jsonStr := content
		if strings.HasPrefix(content, "```json") {
			lines := strings.Split(content, "\n")
			var jsonLines []string
			inJson := false
			for _, line := range lines {
				if strings.HasPrefix(line, "```json") {
					inJson = true
					continue
				}
				if strings.HasPrefix(line, "```") && inJson {
					break
				}
				if inJson {
					jsonLines = append(jsonLines, line)
				}
			}
			jsonStr = strings.Join(jsonLines, "\n")
		}

		if err := json.Unmarshal([]byte(jsonStr), &jsonData); err == nil {
			slog.Debug("Successfully parsed markdown-wrapped JSON", "keys", getJSONKeys(jsonData))

			// Handle structured commit format from markdown-wrapped JSON
			if commitType, hasType := jsonData["type"].(string); hasType {
				if subject, hasSubject := jsonData["subject"].(string); hasSubject {
					message := commitType
					if scope, hasScope := jsonData["scope"].(string); hasScope && scope != "" {
						message += "(" + scope + ")"
					}
					message += ": " + subject

					if body, hasBody := jsonData["body"].(string); hasBody && body != "" {
						message += "\n\n" + body
					}

					if footer, hasFooter := jsonData["footer"].(string); hasFooter && footer != "" {
						message += "\n\n" + footer
					}

					slog.Debug("Built commit message from markdown-wrapped JSON", "message", message)
					return strings.TrimSpace(message)
				}
			}

			// Handle legacy commit_message format in markdown
			if msg, ok := jsonData["commit_message"].(string); ok {
				return strings.TrimSpace(msg)
			}
		}
	}

	// Try to extract from regular lines, looking for conventional commit patterns
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip JSON/markdown formatting lines
		if strings.HasPrefix(line, "{") || strings.HasPrefix(line, "}") ||
			strings.HasPrefix(line, "```") || strings.HasPrefix(line, "\"") ||
			strings.Contains(line, "commit_message") {
			continue
		}

		// Look for conventional commit format: type(scope): description
		if matched, _ := regexp.MatchString(`^(feat|fix|docs|style|refactor|test|chore|build|ci|perf|revert)(\([^)]*\))?: .+`, line); matched {
			return line
		}

		// If we find a non-empty line that looks like a commit message, return it
		if len(line) > 5 && !strings.Contains(line, "{") && !strings.Contains(line, "}") {
			return line
		}
	}

	return ""
}

// getJSONKeys extracts keys from a JSON object for debugging
func getJSONKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

// tryFixTruncatedJSON attempts to fix common truncation issues in JSON responses
func (s *OpenAIService) tryFixTruncatedJSON(content string) string {
	content = strings.TrimSpace(content)

	// If it looks like truncated JSON, try to close it properly
	if strings.HasPrefix(content, "{") && !strings.HasSuffix(content, "}") {
		// Count open/close braces to see if we need to close
		openBraces := strings.Count(content, "{")
		closeBraces := strings.Count(content, "}")

		if openBraces > closeBraces {
			// Try to add missing closing braces and quotes
			fixed := content

			// Handle common truncation patterns
			if strings.HasSuffix(fixed, ",\n  \"") {
				// Truncated in the middle of a field name, remove the incomplete part
				lastCommaIndex := strings.LastIndex(fixed, ",")
				if lastCommaIndex > 0 {
					fixed = fixed[:lastCommaIndex]
				}
			} else if strings.HasSuffix(fixed, ": \"") {
				// Truncated after field name, remove the incomplete value
				lastColonIndex := strings.LastIndex(fixed, ":")
				if lastColonIndex > 0 {
					// Find the start of the field name
					beforeColon := fixed[:lastColonIndex]
					lastQuoteIndex := strings.LastIndex(beforeColon, "\"")
					if lastQuoteIndex > 0 {
						prevQuoteIndex := strings.LastIndex(beforeColon[:lastQuoteIndex], "\"")
						if prevQuoteIndex >= 0 {
							// Remove the incomplete field entirely
							fixed = fixed[:prevQuoteIndex]
							// Remove trailing comma if present
							if strings.HasSuffix(fixed, ",") {
								fixed = fixed[:len(fixed)-1]
							}
						}
					}
				}
			} else if strings.HasSuffix(fixed, "\",\n  \"") {
				// Truncated after a complete field but before the next field name
				lastCommaIndex := strings.LastIndex(fixed, ",")
				if lastCommaIndex > 0 {
					fixed = fixed[:lastCommaIndex]
				}
			}

			// If the last character isn't a quote or brace, add a quote if needed
			if !strings.HasSuffix(fixed, "\"") && !strings.HasSuffix(fixed, "}") && !strings.HasSuffix(fixed, ",") {
				// Check if we're in the middle of a string value
				lastQuoteIndex := strings.LastIndex(fixed, "\"")
				if lastQuoteIndex >= 0 {
					afterLastQuote := fixed[lastQuoteIndex+1:]
					if strings.Contains(afterLastQuote, ":") && !strings.Contains(afterLastQuote, "\"") {
						// We're in a string value, close it
						fixed += "\""
					}
				}
			}

			// Add missing closing braces
			for i := 0; i < (openBraces - closeBraces); i++ {
				fixed += "}"
			}

			slog.Debug("Attempting to fix truncated JSON", "original", content, "fixed", fixed)
			return fixed
		}
	}

	return content
}
