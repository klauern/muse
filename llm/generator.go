package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/templates"
)

type Generator interface {
	Generate(ctx context.Context, diff string, commitStyle templates.CommitStyle) (string, error)
}

type CommitMessageGenerator struct {
	LLMService LLMService
}

func NewCommitMessageGenerator(cfg *config.Config) (*CommitMessageGenerator, error) {
	llmService, err := NewLLMService(&cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return &CommitMessageGenerator{
		LLMService: llmService,
	}, nil
}

func (g *CommitMessageGenerator) Generate(ctx context.Context, diff string, commitStyle templates.CommitStyle) (string, error) {
	fmt.Println("Starting commit message generation")

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempt %d to generate commit message\n", i+1)
		message, err := g.LLMService.GenerateCommitMessage(ctx, diff, commitStyle)
		if err == nil {
			fmt.Printf("Successfully generated commit message: %s\n", message)
			// Attempt to parse the JSON to ensure it's valid
			return message, nil
		} else {
			fmt.Printf("Failed to generate commit message: %v\n", err)
		}

		if i == maxRetries-1 {
			fmt.Printf("Failed to generate valid commit message after %d attempts\n", maxRetries)
			return "", fmt.Errorf("failed to generate valid commit message after %d attempts: %w", maxRetries, err)
		}

		// Wait for a short duration before retrying
		time.Sleep(time.Second * time.Duration(i+1))
	}

	fmt.Println("Unexpected error: should not reach this point")
	return "", fmt.Errorf("unexpected error: should not reach this point")
}
