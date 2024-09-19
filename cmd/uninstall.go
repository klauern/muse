package cmd

import (
	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/core/hook"
	"github.com/urfave/cli/v2"
)

func NewUninstallCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "uninstall",
		Usage: "Uninstall the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			return hook.UninstallHook(config)
		},
	}
}
