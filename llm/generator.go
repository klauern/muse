package llm

import (
	"context"
	"fmt"
	"log/slog"
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
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	llmService, err := NewLLMService(&cfg.LLM)
	if err != nil {
		slog.Error("Failed to create LLM service", "error", err)
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return &CommitMessageGenerator{
		LLMService: llmService,
	}, nil
}

func (g *CommitMessageGenerator) Generate(ctx context.Context, diff string, commitStyle templates.CommitStyle) (string, error) {
	slog.Debug("Generating commit message")

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		slog.Debug("Attempting to generate commit message", "attempt", i+1)
		message, err := g.LLMService.GenerateCommitMessage(ctx, diff, commitStyle)
		if err == nil {
			slog.Debug("Successfully generated commit message", "message", message)
			// Attempt to parse the JSON to ensure it's valid
			return message, nil
		} else {
			slog.Error("Failed to generate commit message", "error", err)
		}

		if i == maxRetries-1 {
			slog.Error("Failed to generate valid commit message after %d attempts", "attempts", maxRetries)
			return "", fmt.Errorf("failed to generate valid commit message after %d attempts: %w", maxRetries, err)
		}

		// Wait for a short duration before retrying
		time.Sleep(time.Second * time.Duration(i+1))
	}

	slog.Error("Unexpected error: should not reach this point")
	return "", fmt.Errorf("unexpected error: should not reach this point")
}
