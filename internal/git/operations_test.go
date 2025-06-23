package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewGitOperations(t *testing.T) {
	// Test with current directory (should be a git repo)
	ops, err := NewGitOperations("")
	if err != nil {
		t.Skipf("Skipping test - not in a git repository: %v", err)
	}

	if ops.workingDir == "" {
		t.Errorf("Expected working directory to be set")
	}

	if ops.timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", ops.timeout)
	}
}

func TestNewGitOperations_InvalidDirectory(t *testing.T) {
	_, err := NewGitOperations("/nonexistent/directory")
	if err == nil {
		t.Errorf("Expected error for nonexistent directory")
	}

	var gitErr GitValidationError
	if !errors.As(err, &gitErr) {
		t.Errorf("Expected GitValidationError, got %T", err)
	}
}

// Helper function to check if we're in a git repository
func isGitRepository() bool {
	ops, err := NewGitOperations("")
	return err == nil && ops != nil
}

func TestValidateGitArgs(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "valid diff command",
			args:      []string{"diff", "--cached"},
			wantError: false,
		},
		{
			name:      "valid status command",
			args:      []string{"status", "--porcelain"},
			wantError: false,
		},
		{
			name:      "invalid command",
			args:      []string{"push", "origin", "main"},
			wantError: true,
		},
		{
			name:      "dangerous exec flag",
			args:      []string{"diff", "--exec=rm -rf /"},
			wantError: true,
		},
		{
			name:      "command injection attempt",
			args:      []string{"diff", "--cached; rm -rf /"},
			wantError: true,
		},
		{
			name:      "shell substitution attempt",
			args:      []string{"diff", "$(rm -rf /)"},
			wantError: true,
		},
		{
			name:      "backtick substitution attempt",
			args:      []string{"diff", "`rm -rf /`"},
			wantError: true,
		},
		{
			name:      "pipe attempt",
			args:      []string{"diff", "--cached | cat"},
			wantError: true,
		},
		{
			name:      "no arguments",
			args:      []string{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ops.validateGitArgs(tt.args)
			if tt.wantError && err == nil {
				t.Errorf("validateGitArgs() expected error for args %v", tt.args)
			}
			if !tt.wantError && err != nil {
				t.Errorf("validateGitArgs() unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}

func TestGetStagedDiff(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	// This should not error even if there are no staged changes
	diff, err := ops.GetStagedDiff()
	if err != nil {
		t.Errorf("GetStagedDiff() error = %v", err)
	}

	// diff can be empty if no staged changes
	if diff == "" {
		t.Logf("No staged changes (empty diff)")
	} else {
		t.Logf("Found staged changes: %d bytes", len(diff))
	}
}

func TestGetRepositoryInfo(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	info, err := ops.GetRepositoryInfo()
	if err != nil {
		t.Errorf("GetRepositoryInfo() error = %v", err)
		return
	}

	if info.Root == "" {
		t.Errorf("Expected repository root to be set")
	}

	if info.Branch == "" {
		t.Errorf("Expected branch name to be set")
	}

	// Verify root is an absolute path
	if !filepath.IsAbs(info.Root) {
		t.Errorf("Expected repository root to be absolute path, got %s", info.Root)
	}

	// Verify root directory exists
	if _, err := os.Stat(info.Root); err != nil {
		t.Errorf("Repository root directory does not exist: %s", info.Root)
	}

	t.Logf("Repository info: root=%s, branch=%s", info.Root, info.Branch)
}

func TestGetStatus(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	status, err := ops.GetStatus()
	if err != nil {
		t.Errorf("GetStatus() error = %v", err)
		return
	}

	// Status can be empty if working tree is clean
	t.Logf("Repository status: %q", status)
}

func TestSetTimeout(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	customTimeout := 5 * time.Second
	ops.SetTimeout(customTimeout)

	if ops.timeout != customTimeout {
		t.Errorf("Expected timeout to be %v, got %v", customTimeout, ops.timeout)
	}
}

func TestExecuteGitCommand_Timeout(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should timeout
	_, err = ops.executeGitCommand(ctx, "status")
	if err == nil {
		t.Errorf("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

func TestGitValidationError(t *testing.T) {
	err := GitValidationError{
		Reason: "test reason",
		Path:   "/test/path",
	}

	expected := "git validation failed: test reason (path: /test/path)"
	if err.Error() != expected {
		t.Errorf("GitValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}

// Test that simulates being outside a git repository
func TestNewGitOperations_NotInGitRepo(t *testing.T) {
	// Create a temporary directory that's not a git repo
	tempDir := t.TempDir()

	_, err := NewGitOperations(tempDir)
	if err == nil {
		t.Errorf("Expected error when not in a git repository")
	}

	var gitErr GitValidationError
	if !errors.As(err, &gitErr) {
		t.Errorf("Expected GitValidationError, got %T", err)
	}

	if !strings.Contains(gitErr.Reason, "not a git repository") {
		t.Errorf("Expected 'not a git repository' error, got: %s", gitErr.Reason)
	}
}

func TestGetStagedDiff_SizeLimit(t *testing.T) {
	if !isGitRepository() {
		t.Skip("Skipping test - not in a git repository")
	}

	ops, err := NewGitOperations("")
	if err != nil {
		t.Fatalf("Failed to create GitOperations: %v", err)
	}

	// This is a bit tricky to test without creating a huge diff
	// But we can at least verify the function doesn't panic
	_, err = ops.GetStagedDiff()
	if err != nil && strings.Contains(err.Error(), "diff too large") {
		t.Logf("Diff size limit triggered: %v", err)
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
