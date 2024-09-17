package config

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	HookConfig HookConfig `mapstructure:"hook"`
	LLM        LLMConfig  `mapstructure:"llm"`
}

// LLMConfig represents the configuration for the LLM service
type LLMConfig struct {
	Provider string                 `mapstructure:"provider"`
	Config   map[string]interface{} `mapstructure:"config"`
}

type HookConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Type        string `mapstructure:"type"`
	CommitStyle string `mapstructure:"commit_style"`
	DryRun      bool   `mapstructure:"dry_run"`
	Preview     bool   `mapstructure:"preview"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("muse")
	v.SetConfigType("yaml")

	// Set default values
	v.SetDefault("hook.enabled", false)
	v.SetDefault("hook.type", "default")
	v.SetDefault("hook.llm_provider", "anthropic") // Default to Anthropic as the LLM provider
	v.SetDefault("hook.commit_style", "default")   // Default commit style

	// Add config search paths
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/muse")
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		v.AddConfigPath(filepath.Join(xdgConfig, "muse"))
	}

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("MUSE")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	decoderConfig := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &config,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(v.AllSettings()); err != nil {
		return nil, err
	}

	return &config, nil
}
