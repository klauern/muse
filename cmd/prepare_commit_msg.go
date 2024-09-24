package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/klauern/muse/config"
	"github.com/klauern/muse/llm"
	"github.com/klauern/muse/rag"
	"github.com/urfave/cli/v2"
)

func NewPrepareCommitMsgCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "prepare-commit-msg",
		Usage: "Run the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			return runPrepareCommitMsg(c, cfg)
		},
	}
}

func runPrepareCommitMsg(c *cli.Context, cfg *config.Config) error {
	if c.NArg() < 1 {
		return fmt.Errorf("missing commit message file argument")
	}

	commitMsgFile := c.Args().Get(0)
	var commitSource string
	if c.NArg() > 1 {
		commitSource = c.Args().Get(1)
	}

	// check the commitSource isn't message, squash, or merge
	if commitSource == "message" || commitSource == "squash" || commitSource == "merge" {
		return nil
	}

	// Get the git diff
	diff, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	// Create RAG service
	ragService := &rag.GitRAGService{}

	// Create commit message generator
	generator, err := llm.NewCommitMessageGenerator(cfg, ragService)
	if err != nil {
		return fmt.Errorf("failed to create commit message generator: %w", err)
	}

	// Generate commit message
	ctx := context.Background()
	message, err := generator.Generate(ctx, diff, cfg.Hook.CommitStyle)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Write the generated message to the commit message file
	if err := os.WriteFile(commitMsgFile, []byte(message), 0o644); err != nil {
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	fmt.Println("Commit message successfully generated and saved.")
	fmt.Println("Prepare commit message hook executed successfully")
	return nil
}
