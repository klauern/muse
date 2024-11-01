package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/templates"
)

// LLMService defines the interface for LLM providers
type LLMService interface {
	GenerateCommitMessage(ctx context.Context, diff string, style templates.CommitStyle) (string, error)
}

// LLMProvider defines the interface for creating LLM services
type LLMProvider interface {
	NewService(config map[string]interface{}) (LLMService, error)
}

var providers = make(map[string]LLMProvider)

// RegisterProvider registers a new LLM provider
func RegisterProvider(name string, provider LLMProvider) {
	providers[name] = provider
}

// NewLLMService creates a new LLMService based on the provided configuration
func NewLLMService(cfg *config.LLMConfig) (LLMService, error) {
	provider, ok := providers[cfg.Provider]
	if !ok {
		slog.Error("Unsupported LLM provider", "provider", cfg.Provider)
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}

	return provider.NewService(cfg.Config)
}
