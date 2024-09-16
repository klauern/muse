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

	// Check for the presence of key components in the commit message
	expectedComponents := []string{
		"feat", // The type of change
		"config", // The scope of the change
		"XDG", // A key term from the change
		":", // Separator in conventional commit format
	}

	for _, component := range expectedComponents {
		if !strings.Contains(commitMessage, component) {
			t.Errorf("Commit message should contain '%s', but it doesn't.\nCommit message: %s", component, commitMessage)
		}
	}

	// Check the structure of the commit message
	lines := strings.Split(strings.TrimSpace(commitMessage), "\n")
	if len(lines) < 2 {
		t.Errorf("Commit message should have at least a subject line and one or more detail lines")
	}

	// Check the subject line (first line)
	subjectLine := strings.SplitN(lines[0], ":", 2)
	if len(subjectLine) != 2 {
		t.Errorf("Subject line should be in the format 'type(scope): description'")
	}

	// Check that the second line is blank
	if len(lines) > 1 && lines[1] != "" {
		t.Errorf("Second line should be blank")
	}

	// Check that there are detail lines
	if len(lines) < 3 {
		t.Errorf("Commit message should include detail lines after the blank line")
	}

	// Check that the commit message doesn't contain the entire diff
	if strings.Contains(commitMessage, "diff --git") {
		t.Errorf("Commit message should not contain the entire diff")
	}

	// Check that the commit message is reasonably sized
	if len(commitMessage) > 2000 {
		t.Errorf("Commit message is too long: %d characters", len(commitMessage))
	}

	// Log the commit message for manual inspection
	t.Logf("Generated commit message:\n%s", commitMessage)
}
