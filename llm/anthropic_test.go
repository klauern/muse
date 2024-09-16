package llm

import (
	"testing"

	"github.com/klauern/pre-commit-llm/config"
)

func TestAnthropicProvider_NewService(t *testing.T) {
	// Save the original ANTHROPIC_API_KEY environment variable
	originalAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	defer os.Setenv("ANTHROPIC_API_KEY", originalAPIKey)

	testCases := []struct {
		name           string
		configAPIKey   string
		envAPIKey      string
		expectedAPIKey string
	}{
		{
			name:           "Config API key",
			configAPIKey:   "config_api_key",
			envAPIKey:      "",
			expectedAPIKey: "config_api_key",
		},
		{
			name:           "Environment API key",
			configAPIKey:   "",
			envAPIKey:      "env_api_key",
			expectedAPIKey: "env_api_key",
		},
		{
			name:           "Config API key takes precedence",
			configAPIKey:   "config_api_key",
			envAPIKey:      "env_api_key",
			expectedAPIKey: "config_api_key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the environment
			os.Setenv("ANTHROPIC_API_KEY", tc.envAPIKey)

			provider := &AnthropicProvider{}
			cfg := &config.LLMConfig{
				Model: "claude-3-sonnet-20240229",
				Extra: map[string]interface{}{},
			}
			if tc.configAPIKey != "" {
				cfg.Extra["api_key"] = tc.configAPIKey
			}

			service, err := provider.NewService(cfg)
			if err != nil {
				t.Fatalf("Failed to create new service: %v", err)
			}

			anthropicService, ok := service.(*AnthropicService)
			if !ok {
				t.Fatalf("Expected AnthropicService, got %T", service)
			}

			if anthropicService.apiKey != tc.expectedAPIKey {
				t.Errorf("Expected API key '%s', got '%s'", tc.expectedAPIKey, anthropicService.apiKey)
			}

			if anthropicService.model != "claude-3-sonnet-20240229" {
				t.Errorf("Expected model 'claude-3-sonnet-20240229', got '%s'", anthropicService.model)
			}
		})
	}
}
