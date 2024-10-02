package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/llm"
	"github.com/urfave/cli/v2"
)

func NewPrepareCommitMsgCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "prepare-commit-msg",
		Usage: "Run the prepare-commit-msg hook",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose logging",
			},
		},
		Action: func(c *cli.Context) error {
			return runPrepareCommitMsg(c, cfg)
		},
	}
}

func runPrepareCommitMsg(c *cli.Context, cfg *config.Config) error {
	verbose := c.Bool("verbose")

	slog.Debug("Verbose mode enabled")

	commitMsgFile, commitSource, err := parseArguments(c)
	if err != nil {
		return err
	}

	slog.Debug("Commit message file", "file", commitMsgFile)
	slog.Debug("Commit source", "source", commitSource)

	if shouldSkipHook(commitSource) {
		slog.Debug("Skipping hook for commit source", "source", commitSource)
		return nil
	}

	diff, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	slog.Debug("Git diff obtained", "length", len(diff))

	message, err := generateCommitMessage(cfg, diff, verbose)
	if err != nil {
		return err
	}

	if err := writeCommitMessage(commitMsgFile, message, verbose); err != nil {
		return err
	}

	slog.Info("Prepare commit message hook executed successfully")
	return nil
}

func parseArguments(c *cli.Context) (string, string, error) {
	if c.NArg() < 1 {
		return "", "", fmt.Errorf("missing commit message file argument")
	}

	commitMsgFile := c.Args().Get(0)
	var commitSource string
	if c.NArg() > 1 {
		commitSource = c.Args().Get(1)
	}

	return commitMsgFile, commitSource, nil
}

func shouldSkipHook(commitSource string) bool {
	return commitSource == "message" || commitSource == "squash" || commitSource == "merge"
}

func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func generateCommitMessage(cfg *config.Config, diff string, verbose bool) (string, error) {
	generator, err := llm.NewCommitMessageGenerator(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create commit message generator: %w", err)
	}

	slog.Debug("Commit message generator created")

	ctx := context.Background()
	message, err := generator.Generate(ctx, diff, cfg.Hook.CommitStyle)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	slog.Debug("Commit message generated successfully")
	slog.Debug("Generated message", "message", message)

	return message, nil
}

func writeCommitMessage(commitMsgFile, message string, verbose bool) error {
	if err := os.WriteFile(commitMsgFile, []byte(message), 0o644); err != nil {
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	slog.Info("Commit message successfully generated and saved.")
	return nil
}
