package cmd

import (
	"github.com/klauern/muse/config"
	"github.com/klauern/muse/core"
	"github.com/urfave/cli/v2"
)

func NewUninstallCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "uninstall",
		Usage: "Uninstall the prepare-commit-msg hook",
		Action: func(c *cli.Context) error {
			installer := core.NewInstaller(config)
			return installer.Uninstall()
		},
	}
}
