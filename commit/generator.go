package commit

import (
	"context"

	"github.com/klauern/pre-commit-llm/llm"
	"github.com/klauern/pre-commit-llm/rag"
)

type CommitMessageGenerator struct {
	LLMService llm.LLMService
	RAGService rag.RAGService
}

func NewCommitMessageGenerator(cfg *config.Config, ragService rag.RAGService) (*CommitMessageGenerator, error) {
	llmClient, err := llm.NewLLMClient(&cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &CommitMessageGenerator{
		LLMService: llmClient,
		RAGService: ragService,
	}, nil
}

func (g *CommitMessageGenerator) Generate(ctx context.Context, diff string) (string, error) {
	context, err := g.RAGService.GetRelevantContext(ctx, diff)
	if err != nil {
		return "", err
	}

	message, err := g.LLMService.GenerateCommitMessage(ctx, diff, context)
	if err != nil {
		return "", err
	}

	return message, nil
}
