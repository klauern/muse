package core

import (
	"context"
	"fmt"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/llm"
	"github.com/klauern/muse/rag"
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

	// Add example commit messages to the context
	examples := `
Examples:
{"type":"feat","scope":"user-auth","subject":"add login functionality","body":"Implemented user login with email and password authentication."}
{"type":"fix","scope":"api","subject":"resolve race condition in data fetching","body":"Fixed a race condition that occurred when multiple requests were made simultaneously to the data fetching endpoint."}
{"type":"docs","scope":"readme","subject":"update installation instructions","body":"Updated the README with clearer installation steps and added troubleshooting section."}
{"type":"refactor","scope":"database","subject":"optimize query performance","body":"Refactored database queries to use indexing, resulting in a 50% reduction in query execution time."}
{"type":"test","scope":"unit-tests","subject":"add tests for user registration","body":"Added comprehensive unit tests to cover all scenarios of the user registration process."}
`
	context = context + "\n" + examples

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
