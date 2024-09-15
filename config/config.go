package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	HookConfig HookConfig
	LLM        LLMConfig
}

type LLMConfig struct {
	Provider         string `mapstructure:"provider"`
	OpenAIAPIKey     string `mapstructure:"openai_api_key"`
	OpenAIModel      string `mapstructure:"openai_model"`
	AnthropicAPIKey  string `mapstructure:"anthropic_api_key"`
	AnthropicModel   string `mapstructure:"anthropic_model"`
	OllamaEndpoint   string `mapstructure:"ollama_endpoint"`
	OllamaModel      string `mapstructure:"ollama_model"`
}

type HookConfig struct {
	Enabled bool
	Type    string
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("muse")
	v.SetConfigType("yaml")

	// Set default values
	v.SetDefault("hook.enabled", false)
	v.SetDefault("hook.type", "default")

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
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
