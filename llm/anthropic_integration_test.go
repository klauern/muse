//go:build integration
// +build integration

package llm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnthropicService_GenerateCommitMessage_Integration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	provider := &AnthropicProvider{}
	cfg := map[string]interface{}{
		"model":   "claude-3-sonnet-20240229",
		"api_key": apiKey,
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

	// Log the generated commit message for debugging and manual inspection
	t.Logf("Generated commit message:\n%s", commitMessage)

	if commitMessage == "" {
		t.Fatal("Generated commit message should not be empty")
	}

	t.Run("CheckComponents", func(t *testing.T) {
		expectedComponents := []string{
			":", // Separator in conventional commit format
		}

		for _, component := range expectedComponents {
			if !strings.Contains(commitMessage, component) {
				t.Errorf("Commit message should contain '%s', but it doesn't", component)
			}
		}
	})

	t.Run("CheckStructure", func(t *testing.T) {
		lines := strings.Split(strings.TrimSpace(commitMessage), "\n")
		if len(lines) < 1 {
			t.Fatal("Commit message should have at least one line")
		}

		if len(lines[0]) > 72 {
			t.Errorf("First line is too long: %d characters (max 72)", len(lines[0]))
		}

		if strings.Contains(commitMessage, "diff --git") {
			t.Errorf("Commit message should not contain the entire diff")
		}

		if strings.Contains(commitMessage, "```") {
			t.Errorf("Commit message should not contain markdown code block markers")
		}
	})

	t.Run("InvalidStyle", func(t *testing.T) {
		_, err := service.GenerateCommitMessage(ctx, string(diffContent), "", CommitStyle(999))
		if err == nil {
			t.Error("Expected error for invalid commit style, but got nil")
		} else if !strings.Contains(err.Error(), "invalid commit style") {
			t.Errorf("Expected error message to contain 'invalid commit style', got: %v", err)
		}
	})
}
