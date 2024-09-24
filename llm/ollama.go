package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type OllamaProvider struct{}

func (p *OllamaProvider) NewService(cfg map[string]interface{}) (LLMService, error) {
	baseURL, ok := cfg["base_url"].(string)
	if !ok || baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama API URL
	}
	model, ok := cfg["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("Ollama model not specified in config")
	}
	return &OllamaService{baseURL: baseURL, model: model}, nil
}

type OllamaService struct {
	baseURL string
	model   string
}

type OllamaRequest struct {
	Model    string        `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Tools    []OllamaTool   `json:"tools,omitempty"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  struct {
			Type       string     `json:"type"`
			Properties Properties `json:"properties"`
			Required   []string   `json:"required"`
		} `json:"parameters"`
	} `json:"function"`
}

type Properties struct {
	Type    Property `json:"type"`
	Scope   Property `json:"scope"`
	Subject Property `json:"subject"`
	Body    Property `json:"body"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type OllamaResponse struct {
	Message OllamaMessage `json:"message"`
}

func (s *OllamaService) GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error) {
	messages := []OllamaMessage{
		{Role: "system", Content: "You are a helpful assistant that generates commit messages based on code diffs."},
		{Role: "user", Content: fmt.Sprintf("Generate a commit message for this diff:\n\n%s\n\nAdditional context:\n%s", diff, context)},
	}

	tools := []OllamaTool{
		{
			Type: "function",
			Function: struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Parameters  struct {
					Type       string     `json:"type"`
					Properties Properties `json:"properties"`
					Required   []string   `json:"required"`
				} `json:"parameters"`
			}{
				Name:        "generate_commit_message",
				Description: "Generate a structured commit message",
				Parameters: struct {
					Type       string     `json:"type"`
					Properties Properties `json:"properties"`
					Required   []string   `json:"required"`
				}{
					Type: "object",
					Properties: Properties{
						Type: Property{
							Type:        "string",
							Description: "The type of change (e.g., feat, fix, docs, style, refactor, test, chore)",
						},
						Scope: Property{
							Type:        "string",
							Description: "The scope of the change (optional)",
						},
						Subject: Property{
							Type:        "string",
							Description: "A short description of the change",
						},
						Body: Property{
							Type:        "string",
							Description: "A more detailed description of the change (optional)",
						},
					},
					Required: []string{"type", "subject"},
				},
			},
		},
	}

	request := OllamaRequest{
		Model:    s.model,
		Messages: messages,
		Stream:   false,
		Tools:    tools,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(s.baseURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request to Ollama API: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var ollamaResp OllamaResponse
	err = json.Unmarshal(body, &ollamaResp)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	return ollamaResp.Message.Content, nil
}

func init() {
	RegisterProvider("ollama", &OllamaProvider{})
}
