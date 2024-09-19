package cmd

import (
	"github.com/klauern/muse/config"
	"github.com/klauern/muse/core"
	"github.com/urfave/cli/v2"
)

func NewConfigureCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "configure",
		Usage: "Generate the configuration file",
		Action: func(c *cli.Context) error {
			return core.CreateConfig()
		},
	}
}
