package main

import (
	"fmt"
	"log"
	"os"

	"github.com/klauern/pre-commit-llm/cmd"
	"github.com/klauern/pre-commit-llm/cmd/muse"
	"github.com/urfave/cli/v2"
)

func main() {
	config, err := muse.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "muse",
		Usage: "A CLI utility for managing git hooks",
		Commands: []*cli.Command{
			cmd.NewStatusCmd(config),
			cmd.NewInstallCmd(config),
			cmd.NewUninstallCmd(config),
			cmd.NewConfigureCmd(config),
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
