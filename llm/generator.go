package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/rag"
)

type Generator interface {
	Generate(ctx context.Context, diff string, commitStyle string) (string, error)
}

type CommitMessageGenerator struct {
	LLMService LLMService
	RAGService rag.RAGService
}

func NewCommitMessageGenerator(cfg *config.Config, ragService rag.RAGService) (*CommitMessageGenerator, error) {
	llmService, err := NewLLMService(&cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM service: %w", err)
	}

	return &CommitMessageGenerator{
		LLMService: llmService,
		RAGService: ragService,
	}, nil
}

func (g *CommitMessageGenerator) Generate(ctx context.Context, diff string, commitStyle string) (string, error) {
	fmt.Println("Starting commit message generation")

	context, err := g.RAGService.GetRelevantContext(ctx, diff)
	if err != nil {
		fmt.Printf("Failed to get relevant context: %v\n", err)
		return "", fmt.Errorf("failed to get relevant context: %w", err)
	}
	fmt.Println("Successfully retrieved relevant context")

	style := GetCommitStyleFromString(commitStyle)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempt %d to generate commit message\n", i+1)
		message, err := g.LLMService.GenerateCommitMessage(ctx, diff, context, style)
		if err == nil {
			fmt.Printf("Successfully generated commit message: %s\n", message)
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
				fmt.Println("Successfully generated and parsed commit message")
				return formattedMessage, nil
			} else {
				fmt.Printf("Failed to parse commit message JSON: %v\n", err)
			}
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
