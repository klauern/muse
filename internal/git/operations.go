package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitOperations provides safe Git operations with validation and security controls
type GitOperations struct {
	workingDir string
	timeout    time.Duration
}

// GitValidationError represents a Git repository validation error
type GitValidationError struct {
	Reason string
	Path   string
}

func (e GitValidationError) Error() string {
	return fmt.Sprintf("git validation failed: %s (path: %s)", e.Reason, e.Path)
}

// NewGitOperations creates a new GitOperations instance with validation
func NewGitOperations(workingDir string) (*GitOperations, error) {
	if workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		workingDir = wd
	}

	ops := &GitOperations{
		workingDir: workingDir,
		timeout:    30 * time.Second,
	}

	if err := ops.validateGitRepository(); err != nil {
		return nil, err
	}

	return ops, nil
}

// SetTimeout configures the timeout for Git operations
func (g *GitOperations) SetTimeout(timeout time.Duration) {
	g.timeout = timeout
}

// validateGitRepository ensures we're in a valid Git repository
func (g *GitOperations) validateGitRepository() error {
	// Check if working directory exists
	if _, err := os.Stat(g.workingDir); os.IsNotExist(err) {
		return GitValidationError{
			Reason: "working directory does not exist",
			Path:   g.workingDir,
		}
	}

	// Check if we're inside a Git repository
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	cmd.Dir = g.workingDir

	output, err := cmd.Output()
	if err != nil {
		return GitValidationError{
			Reason: "not a git repository",
			Path:   g.workingDir,
		}
	}

	gitDir := strings.TrimSpace(string(output))
	if gitDir == "" {
		return GitValidationError{
			Reason: "invalid git directory",
			Path:   g.workingDir,
		}
	}

	// Ensure .git directory is accessible
	var gitPath string
	if filepath.IsAbs(gitDir) {
		gitPath = gitDir
	} else {
		gitPath = filepath.Join(g.workingDir, gitDir)
	}

	if _, err := os.Stat(gitPath); err != nil {
		return GitValidationError{
			Reason: "git directory not accessible",
			Path:   gitPath,
		}
	}

	return nil
}

// executeGitCommand safely executes a Git command with validation
func (g *GitOperations) executeGitCommand(ctx context.Context, args ...string) ([]byte, error) {
	// Validate arguments
	if err := g.validateGitArgs(args); err != nil {
		return nil, fmt.Errorf("invalid git arguments: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.workingDir

	// Set environment to prevent Git from reading user config in some cases
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
	)

	output, err := cmd.Output()
	if err != nil {
		// Enhanced error context
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("git command failed (exit code %d): %s, stderr: %s",
				exitErr.ExitCode(), string(output), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("git command execution failed: %w", err)
	}

	return output, nil
}

// validateGitArgs validates Git command arguments for security
func (g *GitOperations) validateGitArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no git arguments provided")
	}

	// Whitelist allowed git commands for safety
	allowedCommands := map[string]bool{
		"diff":      true,
		"status":    true,
		"rev-parse": true,
		"log":       true,
		"show":      true,
		"branch":    true,
		"config":    true,
	}

	command := args[0]
	if !allowedCommands[command] {
		return fmt.Errorf("git command '%s' not allowed", command)
	}

	// Check for dangerous flags
	for _, arg := range args {
		if strings.HasPrefix(arg, "--exec=") ||
			strings.HasPrefix(arg, "--upload-pack=") ||
			strings.HasPrefix(arg, "--receive-pack=") ||
			strings.Contains(arg, "$(") ||
			strings.Contains(arg, "`") ||
			strings.Contains(arg, "|") ||
			strings.Contains(arg, "&") ||
			strings.Contains(arg, ";") {
			return fmt.Errorf("dangerous git argument: %s", arg)
		}
	}

	return nil
}

// GetStagedDiff safely retrieves the staged changes
func (g *GitOperations) GetStagedDiff() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	output, err := g.executeGitCommand(ctx, "diff", "--cached", "--no-ext-diff")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	diff := string(output)

	// Validate diff size to prevent memory exhaustion
	const maxDiffSize = 1024 * 1024 // 1MB
	if len(diff) > maxDiffSize {
		return "", fmt.Errorf("diff too large (%d bytes), maximum allowed: %d bytes",
			len(diff), maxDiffSize)
	}

	return diff, nil
}

// GetRepository returns information about the current repository
func (g *GitOperations) GetRepositoryInfo() (*RepositoryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Get repository root
	rootOutput, err := g.executeGitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	// Get current branch
	branchOutput, err := g.executeGitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	return &RepositoryInfo{
		Root:   strings.TrimSpace(string(rootOutput)),
		Branch: strings.TrimSpace(string(branchOutput)),
	}, nil
}

// GetStatus safely retrieves the repository status
func (g *GitOperations) GetStatus() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	output, err := g.executeGitCommand(ctx, "status", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %w", err)
	}

	return string(output), nil
}

// RepositoryInfo contains basic repository information
type RepositoryInfo struct {
	Root   string
	Branch string
}
