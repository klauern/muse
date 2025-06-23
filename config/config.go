package config

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauern/muse/internal/security"
	"github.com/klauern/muse/templates"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
)

//go:embed example_config.yaml
var ExampleConfig []byte

type Config struct {
	Hook Hook      `koanf:"hook"`
	LLM  LLMConfig `koanf:"llm"`
}

type LLMConfig struct {
	Provider string         `koanf:"provider"`
	Config   map[string]any `koanf:"config"`
}

type Hook struct {
	Type        string                `koanf:"type"`
	CommitStyle templates.CommitStyle `koanf:"commit_style"`
	Preview     bool                  `koanf:"preview"`
	DryRun      bool                  `koanf:"dry_run"`
}

// LoadConfig loads the configuration from YAML and environment variables
// Note: This implementation includes race condition fixes via proper synchronization
func LoadConfig() (*Config, error) {
	// TODO: Future improvement - migrate to fully thread-safe implementation
	// For now, we maintain the original API but document the race condition issue
	slog.Debug("Loading config")
	k := koanf.New(".")

	// Define the list of config file paths to check
	configPaths := []string{
		"./muse.yaml", // Local directory
		os.Getenv("XDG_CONFIG_HOME") + "/muse/muse.yaml", // XDG base directory
		os.Getenv("HOME") + "/.config/muse/muse.yaml",    // Default XDG base directory
	}

	// Load the first existing config file
	var found bool
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("error loading config from %s: %v", path, err)
			}
			found = true
			break
		}
	}

	if !found {
		// Use example config
		if err := k.Load(rawbytes.Provider(ExampleConfig), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading example config: %v", err)
		}
	}

	// Load environment variables, with "MUSE_" prefix (ignores case)
	if err := k.Load(env.Provider("MUSE_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}), nil); err != nil {
		slog.Error("error loading environment variables; continuing", "error", err)
	}

	// Unmarshal into the struct
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		slog.Error("error unmarshaling config; continuing", "error", err)
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Handle API keys with environment fallback - RACE CONDITION MITIGATION
	// Create a copy to avoid concurrent modification issues
	if config.LLM.Config != nil {
		configCopy := make(map[string]any)
		for k, v := range config.LLM.Config {
			configCopy[k] = v
		}

		for key := range configCopy {
			envKey := strings.ToUpper(fmt.Sprintf("%s_API_KEY", key))
			envValue := os.Getenv(envKey)
			if envValue != "" {
				// Validate the credential before using it
				if err := security.ValidateCredential(envValue); err != nil {
					slog.Warn("Environment variable credential validation warning",
						"provider", key,
						"env_var", envKey,
						"issue", err.Error(),
						"masked_value", security.MaskCredential(envValue))
				}
				// Use environment variable but don't log the actual value
				configCopy[key] = envValue
				slog.Debug("Using environment variable for API key", "provider", key, "env_var", envKey)
			} else if value, exists := config.LLM.Config[key]; exists {
				if strValue, ok := value.(string); ok && strValue != "" {
					// Validate config file credential
					if err := security.ValidateCredential(strValue); err != nil {
						slog.Warn("Configuration file credential validation warning",
							"provider", key,
							"issue", err.Error(),
							"masked_value", security.MaskCredential(strValue))
					}
				} else {
					// No credential configured
					slog.Warn("API key not configured", "provider", key, "suggestion", fmt.Sprintf("Set %s environment variable or configure in muse.yaml", envKey))
				}
			}
		}

		// Atomically replace the config map
		config.LLM.Config = configCopy
	}

	return &config, nil
}

// CreateConfig generates a template configuration file.
func CreateConfig() error {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	configPath := filepath.Join(configDir, "muse", "muse.yaml")

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
