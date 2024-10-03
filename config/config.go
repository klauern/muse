package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

//go:embed example_config.yaml
var ExampleConfig []byte

type Config struct {
	Hook Hook      `koanf:"hook"`
	LLM  LLMConfig `koanf:"llm"`
}

type LLMConfig struct {
	Provider string                 `koanf:"provider"`
	Config   map[string]interface{} `koanf:"config"`
}

type Hook struct {
	Type        string `koanf:"type"`
	CommitStyle string `koanf:"commit_style"`
	Preview     bool   `koanf:"preview"`
	DryRun      bool   `koanf:"dry_run"`
}

// LoadConfig loads the configuration from YAML and environment variables
func LoadConfig() (*Config, error) {
	k := koanf.New(".")

	// Load YAML config file
	if err := k.Load(file.Provider("muse.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	// Load environment variables, with "MUSE_" prefix (ignores case)
	k.Load(env.Provider("MUSE_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	}), nil)

	// Unmarshal into the struct
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Handle API keys with environment fallback
	for key, value := range config.LLM.Config {
		envKey := strings.ToUpper(fmt.Sprintf("%s_API_KEY", key))
		envValue := os.Getenv(envKey)
		if envValue != "" {
			config.LLM.Config[key] = envValue
		} else if strValue, ok := value.(string); ok && strValue == "" {
			config.LLM.Config[key] = "your-default-api-key" // final fallback
		}
	}

	return &config, nil
}

// CreateConfig generates a template configuration file.
func CreateConfig() error {
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
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	if err := os.WriteFile(configPath, ExampleConfig, 0o644); err != nil {
		return fmt.Errorf("failed to write template config: %w", err)
	}

	fmt.Printf("Template configuration file generated at %s\n", configPath)
	return nil
}
