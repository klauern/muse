package commit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/llm"
	"github.com/klauern/pre-commit-llm/rag"
)

type Generator interface {
	Generate(ctx context.Context, diff string, commitStyle string) (string, error)
}

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
			// Attempt to parse the JSON to ensure it's valid
			var parsedMessage struct {
				Type    string `json:"type"`
				Scope   string `json:"scope"`
				Subject string `json:"subject"`
				Body    string `json:"body"`
			}
			if err := json.Unmarshal([]byte(message), &parsedMessage); err == nil {
				// Format the commit message
				formattedMessage := fmt.Sprintf("%s(%s): %s\n\n%s", 
					parsedMessage.Type, 
					parsedMessage.Scope, 
					parsedMessage.Subject, 
					parsedMessage.Body)
				return formattedMessage, nil
			}
		}
		
		if i == maxRetries-1 {
			return "", fmt.Errorf("failed to generate valid commit message after %d attempts: %w", maxRetries, err)
		}
		
		// Wait for a short duration before retrying
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return "", fmt.Errorf("unexpected error: should not reach this point")
}
