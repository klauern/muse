//go:build integration
// +build integration

package llm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauern/pre-commit-llm/config"
)

func TestAnthropicService_GenerateCommitMessage_Integration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	provider := &AnthropicProvider{}
	cfg := &config.LLMConfig{
		Model: "claude-3-sonnet-20240229",
		Extra: map[string]interface{}{
			"api_key": apiKey,
		},
	}

	service, err := provider.NewService(cfg)
	if err != nil {
		t.Fatalf("Failed to create new service: %v", err)
	}

	// Read the diff file
	diffPath := filepath.Join("..", "testdata", "diffs", "small.diff")
	diffContent, err := os.ReadFile(diffPath)
	if err != nil {
		t.Fatalf("Failed to read diff file: %v", err)
	}

	ctx := context.Background()
	commitMessage, err := service.GenerateCommitMessage(ctx, string(diffContent), "", ConventionalStyle)
	if err != nil {
		t.Fatalf("Failed to generate commit message: %v", err)
	}

	// Print the generated commit message for debugging
	t.Logf("Generated commit message: %s", commitMessage)

	if commitMessage == "" {
		t.Error("Generated commit message should not be empty")
	}

	if !strings.Contains(strings.ToLower(commitMessage), "xdg") {
		t.Error("Commit message should mention XDG specification")
	}
}
