package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/briandowns/spinner"
	"github.com/klauern/muse/config"
	"github.com/klauern/muse/llm"
	"github.com/urfave/cli/v2"
)

func NewPrepareCommitMsgCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "prepare-commit-msg",
		Usage: "Run the prepare-commit-msg hook or generate a commit message",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose logging",
			},
			&cli.BoolFlag{
				Name:  "generate",
				Usage: "Generate and print a commit message without writing to a file",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("verbose") {
				slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
			}
			return runPrepareCommitMsg(c, cfg)
		},
	}
}

func runPrepareCommitMsg(c *cli.Context, cfg *config.Config) error {
	generateOnly := c.Bool("generate")

	slog.Debug("Verbose mode enabled")

	if generateOnly {
		return generateAndPrintCommitMessage(cfg)
	}

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

	message, err := generateCommitMessage(cfg, diff)
	if err != nil {
		return err
	}

	if err := writeCommitMessage(commitMsgFile, message); err != nil {
		return err
	}

	slog.Info("Prepare commit message hook executed successfully")
	return nil
}

func generateAndPrintCommitMessage(cfg *config.Config) error {
	diff, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	slog.Debug("Git diff obtained", "length", len(diff))

	message, err := generateCommitMessage(cfg, diff)
	if err != nil {
		return err
	}

	slog.Info("Generated commit message", "message", message)

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
	slog.Debug("Executing git diff command")
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to execute git diff command", "error", err)
		return "", err
	}
	slog.Debug("Git diff command executed successfully", "output_length", len(output))
	return string(output), nil
}

func generateCommitMessage(cfg *config.Config, diff string) (string, error) {
	slog.Debug("Starting commit message generation")
	generator, err := llm.NewCommitMessageGenerator(cfg)
	if err != nil {
		slog.Error("Failed to create commit message generator", "error", err)
		return "", fmt.Errorf("failed to create commit message generator: %w", err)
	}

	slog.Debug("Commit message generator created successfully")
	ctx := context.Background()
	slog.Debug("Generating commit message", "diff_length", len(diff), "commit_style", cfg.Hook.CommitStyle)

	// Create and start the spinner
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Generating commit message..."
	s.Start()

	message, err := generator.Generate(ctx, diff, cfg.Hook.CommitStyle)

	// Stop the spinner
	s.Stop()

	if err != nil {
		slog.Error("Failed to generate commit message", "error", err)
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	slog.Debug("Commit message generated successfully", "message_length", len(message))
	return message, nil
}

func writeCommitMessage(commitMsgFile, message string) error {
	if err := os.WriteFile(commitMsgFile, []byte(message), 0o644); err != nil {
		slog.Error("Failed to write commit message", "error", err)
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	slog.Info("Commit message successfully generated and saved.")
	return nil
}
