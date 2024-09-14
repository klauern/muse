package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yourusername/muse"
)

func NewConfigureCmd(config *muse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure the prepare-commit-msg hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configureHook(config)
		},
	}

	cmd.Flags().Bool("enabled", config.HookConfig.Enabled, "Enable or disable the hook")
	cmd.Flags().String("type", config.HookConfig.Type, "Set the hook type")

	return cmd
}

func configureHook(config *muse.Config) error {
	v := viper.GetViper()

	enabled, _ := v.Get("enabled").(bool)
	hookType, _ := v.Get("type").(string)

	v.Set("hook.enabled", enabled)
	v.Set("hook.type", hookType)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Configuration updated: Enabled=%v, Type=%s\n", enabled, hookType)
	return nil
}
