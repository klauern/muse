package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/muse"
	"github.com/yourusername/muse/cmd"
)

func main() {
	config, err := muse.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "muse",
		Short: "Muse is a CLI utility for managing git hooks",
		Long:  `Muse allows you to manage and configure the prepare-commit-msg git hook.`,
	}

	rootCmd.AddCommand(
		cmd.NewStatusCmd(config),
		cmd.NewInstallCmd(config),
		cmd.NewUninstallCmd(config),
		cmd.NewConfigureCmd(config),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
