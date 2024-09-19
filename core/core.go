package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/llm"
	"github.com/klauern/pre-commit-llm/rag"
)

// CommitMessageGenerator struct
type CommitMessageGenerator struct {
	LLMService llm.LLMService
	RAGService rag.RAGService
}

func NewCommitMessageGenerator(cfg *config.Config, ragService rag.RAGService) (*CommitMessageGenerator, error) {
	llmService, err := llm.NewLLMService(&cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return &CommitMessageGenerator{
		LLMService: llmService,
		RAGService: ragService,
	}, nil
}

func (g *CommitMessageGenerator) Generate(ctx context.Context, diff string, commitStyle string) (string, error) {
	context, err := g.RAGService.GetRelevantContext(ctx, diff)
	if err != nil {
		return "", fmt.Errorf("failed to get relevant context: %w", err)
	}

	style := llm.GetCommitStyleFromString(commitStyle)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		message, err := g.LLMService.GenerateCommitMessage(ctx, diff, context, style)
		if err == nil {
			return message, nil
		}
		// Log error and retry
	}

	return "", fmt.Errorf("failed to generate commit message after %d attempts", maxRetries)
}

// CreateConfig generates a template configuration file.
func CreateConfig() error {
	// Determine the configuration directory based on the XDG specification
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	configPath := filepath.Join(configDir, "muse", "muse.yaml")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	if err := os.WriteFile(configPath, config.ExampleConfig, 0o644); err != nil {
		return fmt.Errorf("failed to write template config: %w", err)
	}

	fmt.Printf("Template configuration file generated at %s\n", configPath)
	return nil
}

// Installer struct
type Installer struct {
	config *config.Config
}

func NewInstaller(config *config.Config) *Installer {
	return &Installer{config: config}
}

func (i *Installer) Install() error {
	gitDir, err := FindGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		return fmt.Errorf("prepare-commit-msg hook already exists")
	}

	// Create the hook file
	f, err := os.Create(hookPath)
	if err != nil {
		return fmt.Errorf("failed to create hook file: %w", err)
	}
	defer f.Close()

	// Write the hook content
	hookContent := `#!/bin/sh
exec < /dev/tty
muse prepare-commit-msg "$1" "$2" "$3"
`
	if _, err := f.WriteString(hookContent); err != nil {
		return fmt.Errorf("failed to write hook content: %w", err)
	}

	// Make the hook executable
	if err := os.Chmod(hookPath, 0o755); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	fmt.Println("prepare-commit-msg hook installed successfully")
	return nil
}

func (i *Installer) Uninstall() error {
	gitDir, err := FindGitDir()
	if err != nil {
		return fmt.Errorf("failed to find .git directory: %w", err)
	}

	hookPath := filepath.Join(gitDir, "hooks", "prepare-commit-msg")

	// Check if hook exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return fmt.Errorf("prepare-commit-msg hook does not exist")
	}

	// Remove the hook file
	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove hook file: %w", err)
	}

	fmt.Println("prepare-commit-msg hook uninstalled successfully")
	return nil
}

func FindGitDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return filepath.Join(dir, ".git"), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}
