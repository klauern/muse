package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/klauern/muse/cmd"
	"github.com/klauern/muse/config"
	"github.com/urfave/cli/v2"
)

func main() {
	var cfg *config.Config

	app := &cli.App{
		Name:    "muse",
		Usage:   "A CLI utility for managing git hooks",
		Version: fmt.Sprintf("%s (commit: %s, built at: %s)", cmd.Version, cmd.CommitHash, cmd.BuildDate),
		Commands: []*cli.Command{
			cmd.NewStatusCmd(cfg),
			cmd.NewInstallCmd(cfg),
			cmd.NewUninstallCmd(cfg),
			cmd.NewConfigureCmd(cfg),
			cmd.NewPrepareCommitMsgCmd(cfg),
			{
				Name:  "version",
				Usage: "Print the version",
				Action: func(c *cli.Context) error {
					fmt.Printf("muse version %s\ncommit: %s\nbuilt at: %s\n", cmd.Version, cmd.CommitHash, cmd.BuildDate)
					return nil
				},
			},
		},
		Before: func(c *cli.Context) error {
			if c.Bool("verbose") {
				slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
			}

			var err error
			cfg, err = config.LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		slog.Error("Error running app", "error", err)
		os.Exit(1)
	}
}
