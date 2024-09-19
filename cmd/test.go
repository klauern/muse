package cmd

import (
	"context"
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/llm"
	"github.com/urfave/cli/v2"
)

func NewTestCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "test",
		Usage: "Test the LLM service connection",
		Action: func(c *cli.Context) error {
			return testLLMService(cfg)
		},
	}
}

func testLLMService(cfg *config.Config) error {
	// Ensure the provider is set to Anthropic
	cfg.LLM.Provider = "anthropic"

	// Create LLM service
	llmService, err := llm.NewLLMService(&cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create LLM service: %w", err)
	}

	// Test the service with a simple prompt
	ctx := context.Background()
	testDiff := "This is a test diff"
	testContext := "This is a test context"
	response, err := llmService.GenerateCommitMessage(ctx, testDiff, testContext, llm.DefaultStyle)
	if err != nil {
		return fmt.Errorf("failed to generate test message: %w", err)
	}

	fmt.Println("Test successful! Response from Anthropic API:")
	fmt.Println(response)

	fmt.Println("LLM service test completed successfully")
	return nil
}
