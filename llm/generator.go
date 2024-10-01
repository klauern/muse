package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/klauern/muse/config"
)

type Generator interface {
	Generate(ctx context.Context, diff string, commitStyle string) (string, error)
}

type CommitMessageGenerator struct {
	LLMService LLMService
}

type GeneratedCommitMessage struct {
	Type    string `json:"type" jsonschema:"title=Type of commit message,description=Type of commit message,enum=feat,enum=fix,enum=chore,enum=docs,enum=style,enum=refactor,enum=perf,enum=test,enum=ci,enum=build,enum=release,required=true"`
	Scope   string `json:"scope,omitempty" jsonschema:"title=Scope of commit message,description=Scope of commit message,optional=true"`
	Subject string `json:"subject" jsonschema:"title=Subject of commit message,description=Subject of commit message,maxLength=72,required=true"`
	Body    string `json:"body" jsonschema:"title=Body of commit message,description=Detailed description of commit message,optional=true"`
}

func (g GeneratedCommitMessage) String() string {
	return fmt.Sprintf("%s(%s): %s\n\n%s", g.Type, g.Scope, g.Subject, g.Body)
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

func (g *CommitMessageGenerator) Generate(ctx context.Context, diff string, commitStyle string) (string, error) {
	fmt.Println("Starting commit message generation")

	style := GetCommitStyleFromString(commitStyle)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempt %d to generate commit message\n", i+1)
		message, err := g.LLMService.GenerateCommitMessage(ctx, diff, style)
		if err == nil {
			fmt.Printf("Successfully generated commit message: %s\n", message)
			// Attempt to parse the JSON to ensure it's valid
			var parsedMessage GeneratedCommitMessage
			if err := json.Unmarshal([]byte(message), &parsedMessage); err == nil {
				fmt.Println("Successfully generated and parsed commit message")
				return parsedMessage.String(), nil
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
