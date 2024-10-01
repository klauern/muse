package cmd

import (
	"context"
	"fmt"
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

	if verbose {
		fmt.Println("Verbose mode enabled")
	}

	if c.NArg() < 1 {
		return fmt.Errorf("missing commit message file argument")
	}

	commitMsgFile := c.Args().Get(0)
	var commitSource string
	if c.NArg() > 1 {
		commitSource = c.Args().Get(1)
	}

	if verbose {
		fmt.Printf("Commit message file: %s\n", commitMsgFile)
		fmt.Printf("Commit source: %s\n", commitSource)
	}

	// check the commitSource isn't message, squash, or merge
	if commitSource == "message" || commitSource == "squash" || commitSource == "merge" {
		if verbose {
			fmt.Printf("Skipping hook for commit source: %s\n", commitSource)
		}
		return nil
	}

	// Get the git diff
	diff, err := getGitDiff()
	if err != nil {
		return fmt.Errorf("failed to get git diff: %w", err)
	}

	if verbose {
		fmt.Printf("Git diff obtained, length: %d characters\n", len(diff))
	}

	// Create commit message generator
	generator, err := llm.NewCommitMessageGenerator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create commit message generator: %w", err)
	}

	if verbose {
		fmt.Println("Commit message generator created")
	}

	// Generate commit message
	ctx := context.Background()
	message, err := generator.Generate(ctx, diff, cfg.Hook.CommitStyle)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	if verbose {
		fmt.Println("Commit message generated successfully")
		fmt.Printf("Generated message:\n%s\n", message)
	}

	// Write the generated message to the commit message file
	if err := os.WriteFile(commitMsgFile, []byte(message), 0o644); err != nil {
		return fmt.Errorf("failed to write commit message: %w", err)
	}

	if verbose {
		fmt.Println("Commit message successfully written to file")
	}

	fmt.Println("Commit message successfully generated and saved.")
	fmt.Println("Prepare commit message hook executed successfully")
	return nil
}

func getGitDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
