package cmd

import (
	"github.com/klauern/muse/config"
	"github.com/klauern/muse/hooks"
	"github.com/urfave/cli/v2"
)

func NewInstallCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			installer := hooks.NewInstaller(config)
			return installer.Install()
		},
	}
}
