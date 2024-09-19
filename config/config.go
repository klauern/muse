package config

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

//go:embed example_config.yaml
var ExampleConfig []byte

type Config struct {
	Hook Hook      `mapstructure:"hook"`
	LLM  LLMConfig `mapstructure:"llm"`
}

type LLMConfig struct {
	Provider string                 `mapstructure:"provider"`
	Config   map[string]interface{} `mapstructure:"config"`
}

type Hook struct {
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

	setDefaults(v)
	setConfigPaths(v)

	v.AutomaticEnv()
	v.SetEnvPrefix("MUSE")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("hook.llm_provider", "anthropic")
	v.SetDefault("hook.commit_style", "default")
}

func setConfigPaths(v *viper.Viper) {
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/muse")
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		v.AddConfigPath(filepath.Join(xdgConfig, "muse"))
	}
}
