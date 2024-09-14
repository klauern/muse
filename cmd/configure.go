package cmd

import (
	"fmt"

	"github.com/klauern/pre-commit-llm/config"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func NewConfigureCmd(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "configure",
		Usage: "Configure the prepare-commit-msg hook",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "enabled",
				Usage: "Enable or disable the hook",
				Value: config.HookConfig.Enabled,
			},
			&cli.StringFlag{
				Name:  "type",
				Usage: "Set the hook type",
				Value: config.HookConfig.Type,
			},
		},
		Action: func(c *cli.Context) error {
			return configureHook(c, config)
		},
	}
}

func configureHook(c *cli.Context, config *config.Config) error {
	v := viper.GetViper()

	enabled := c.Bool("enabled")
	hookType := c.String("type")

	v.Set("hook.enabled", enabled)
	v.Set("hook.type", hookType)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Configuration updated: Enabled=%v, Type=%s\n", enabled, hookType)
	return nil
}
