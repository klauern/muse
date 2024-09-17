package llm

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/templates"
)

// CommitStyle represents different commit message styles
type CommitStyle int

const (
	DefaultStyle CommitStyle = iota
	ConventionalStyle
	GitmojisStyle
)

func (cs CommitStyle) String() string {
	switch cs {
	case ConventionalStyle:
		return "conventional"
	case GitmojisStyle:
		return "gitmojis"
	default:
		return "default"
	}
}

// LLMService defines the interface for LLM providers
type LLMService interface {
	GenerateCommitMessage(ctx context.Context, diff, context string, style CommitStyle) (string, error)
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

// GetCommitTemplate returns the appropriate template based on the commit style
func GetCommitTemplate(style CommitStyle) *template.Template {
	switch style {
	case ConventionalStyle:
		return templates.ConventionalCommitTemplate
	case GitmojisStyle:
		return templates.GitmojisTemplate
	default:
		return templates.DefaultCommitTemplate
	}
}

// GetCommitStyleFromString converts a string representation of commit style to CommitStyle enum
func GetCommitStyleFromString(style string) CommitStyle {
	switch strings.ToLower(style) {
	case "conventional":
		return ConventionalStyle
	case "gitmojis":
		return GitmojisStyle
	default:
		return DefaultStyle
	}
}
