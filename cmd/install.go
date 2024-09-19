package cmd

import (
	"github.com/klauern/pre-commit-llm/config"
	"github.com/klauern/pre-commit-llm/core/hook"
	"github.com/urfave/cli/v2"
)

func NewInstallCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			installer := hook.NewInstaller(config)
			return installer.Install()
		},
	}
}
