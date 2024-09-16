package llm

import (
	"context"
	"testing"

	"github.com/klauern/pre-commit-llm/config"
)

func TestAnthropicProvider_NewService(t *testing.T) {
	provider := &AnthropicProvider{}
	cfg := &config.LLMConfig{
		Model: "claude-3-sonnet-20240229",
		Extra: map[string]interface{}{
			"api_key": "test_api_key",
		},
	}

	service, err := provider.NewService(cfg)
	if err != nil {
		t.Fatalf("Failed to create new service: %v", err)
	}

	anthropicService, ok := service.(*AnthropicService)
	if !ok {
		t.Fatalf("Expected AnthropicService, got %T", service)
	}

	if anthropicService.apiKey != "test_api_key" {
		t.Errorf("Expected API key 'test_api_key', got '%s'", anthropicService.apiKey)
	}

	if anthropicService.model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected model 'claude-3-sonnet-20240229', got '%s'", anthropicService.model)
	}
}

// Add more unit tests here as needed
