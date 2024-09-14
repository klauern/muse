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
