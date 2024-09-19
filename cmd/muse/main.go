package main

import (
	"fmt"
	"log"
	"os"

	"github.com/klauern/pre-commit-llm/cmd"
	"github.com/klauern/pre-commit-llm/config"
	"github.com/urfave/cli/v2"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:    "muse",
		Usage:   "A CLI utility for managing git hooks",
		Version: cmd.Version,
		Commands: []*cli.Command{
			cmd.NewStatusCmd(cfg),
			cmd.NewInstallCmd(cfg),
			cmd.NewUninstallCmd(cfg),
			cmd.NewConfigureCmd(cfg),
			cmd.NewGenerateCmd(cfg),
			cmd.NewPrepareCommitMsgCmd(cfg),
			cmd.NewTestCmd(cfg),
			{
				Name:  "version",
				Usage: "Print the version",
				Action: func(c *cli.Context) error {
					fmt.Printf("muse version %s\n", cmd.Version)
					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
