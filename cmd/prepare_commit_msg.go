package cmd

import (
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/hooks"
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
	commitSource := c.Args().Get(1)
	sha1 := c.Args().Get(2)

	hook, err := hooks.NewHook(cfg)
	if err != nil {
		return fmt.Errorf("failed to create hook: %w", err)
	}

	return hook.Run(commitMsgFile, commitSource, sha1)
}
