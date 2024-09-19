package cmd

import (
	"fmt"
	"os/exec"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/core/generator"
	"github.com/klauern/pre-commit-llm/rag"
	"github.com/urfave/cli/v2"
)

func NewGenerateCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate a commit message using a specific LLM provider",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "LLM provider to use (e.g., anthropic, openai)",
				Value: cfg.LLM.Provider,
			},
			&cli.StringFlag{
				Name:  "style",
				Usage: "Commit message style (default, conventional, gitmojis)",
				Value: cfg.Hook.CommitStyle,
			},
		},
		Action: func(c *cli.Context) error {
			return generateCommitMessage(c, cfg)
		},
	}
}

func generateCommitMessage(c *cli.Context, cfg *config.Config) error {
	provider := c.String("provider")
	style := c.String("style")

	// Create a copy of the LLM config and override the provider
	llmConfig := cfg.LLM
	llmConfig.Provider = provider

	// Create RAG service
	ragService := &rag.GitRAGService{}

	// Create commit message generator
	gen, err := generator.NewCommitMessageGenerator(cfg, ragService)
	if err != nil {
		return fmt.Errorf("failed to create commit message generator: %w", err)
	}

	// Get the git diff
	diff, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Generate commit message
	message, err := gen.Generate(c.Context, diff, style)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Print the generated message
	fmt.Printf("Generated commit message using %s provider:\n\n%s\n", provider, message)

	fmt.Println("Commit message generated successfully")
	return nil
}

func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
