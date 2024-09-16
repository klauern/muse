package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
			&cli.BoolFlag{
				Name:  "template",
				Usage: "Generate a template configuration file",
			},
		},
		Action: func(c *cli.Context) error {
			return configureHook(c, config)
		},
	}
}

func configureHook(c *cli.Context, config *config.Config) error {
	if c.Bool("template") {
		return generateTemplateConfig()
	}

	v := viper.GetViper()

	if c.IsSet("enabled") {
		enabled := c.Bool("enabled")
		v.Set("hook.enabled", enabled)
		fmt.Printf("Configuration updated: Enabled=%v\n", enabled)
	}

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func generateTemplateConfig() error {
	exampleConfig, err := os.ReadFile("config/example_config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read example config: %w", err)
	}

	// Determine the configuration directory based on the XDG specification
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	configPath := filepath.Join(configDir, "muse", "muse.yaml")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	if err := os.WriteFile(configPath, exampleConfig, 0644); err != nil {
		return fmt.Errorf("failed to write template config: %w", err)
	}

	fmt.Printf("Template configuration file generated at %s\n", configPath)
	return nil
}
