package generator

import (
	"context"
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/llm"
	"github.com/klauern/pre-commit-llm/rag"
)

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
