package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/templates"
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
	GenerateCommitMessage(ctx context.Context, diff string, style CommitStyle) (string, error)
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

	if cfg.Provider == "ollama" {
		return provider.NewService(cfg.Config)
	}

	return provider.NewService(cfg.Config)
}

// GetCommitTemplate returns the appropriate template based on the commit style
func GetCommitTemplate(style CommitStyle) templates.CommitTemplate {
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

type GeneratedCommitMessage struct {
	Type    string `json:"type" jsonschema:"title=Type of commit message,description=Type of commit message,enum=feat,enum=fix,enum=chore,enum=docs,enum=style,enum=refactor,enum=perf,enum=test,enum=ci,enum=build,enum=release,required=true"`
	Scope   string `json:"scope,omitempty" jsonschema:"title=Scope of commit message,description=Scope of commit message,optional=true"`
	Subject string `json:"subject" jsonschema:"title=Subject of commit message,description=Subject of commit message,maxLength=72,required=true"`
	Body    string `json:"body" jsonschema:"title=Body of commit message,description=Detailed description of commit message,optional=true"`
}

func (g GeneratedCommitMessage) String() string {
	return fmt.Sprintf("%s(%s): %s\n\n%s", g.Type, g.Scope, g.Subject, g.Body)
}
