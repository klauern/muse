package llm

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}

	return provider.NewService(cfg.Config)
}

// // GetCommitTemplate returns the appropriate template based on the commit style
// func GetCommitTemplate(style templates.CommitStyle, tm *templates.TemplateManager) templates.CommitTemplate {
// 	switch style {
// 	case templates.ConventionalCommitStyle:
// 		return tm.ConventionalCommit
// 	case templates.GitmojiCommitStyle:
// 		return tm.Gitmojis
// 	default:
// 		return tm.DefaultCommit
// 	}
// }
