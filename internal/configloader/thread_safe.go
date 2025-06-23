package configloader

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	museconfig "github.com/klauern/muse/config"
	"github.com/klauern/muse/internal/security"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
)

// ConfigLoader provides thread-safe configuration loading with caching
type ConfigLoader struct {
	mu           sync.RWMutex
	cachedConfig *CachedConfig
	loadTimeout  time.Duration
}

// CachedConfig represents a cached configuration with expiration
type CachedConfig struct {
	Config    *museconfig.Config
	LoadTime  time.Time
	ExpiresAt time.Time
}

// ConfigError represents a configuration loading error
type ConfigError struct {
	Stage  string
	Path   string
	Reason string
	Err    error
}

func (e ConfigError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("config %s failed for %s: %s (%v)", e.Stage, e.Path, e.Reason, e.Err)
	}
	return fmt.Sprintf("config %s failed: %s (%v)", e.Stage, e.Reason, e.Err)
}

func (e ConfigError) Unwrap() error {
	return e.Err
}

var (
	globalLoader *ConfigLoader
	loaderOnce   sync.Once
)

// GetConfigLoader returns the singleton configuration loader
func GetConfigLoader() *ConfigLoader {
	loaderOnce.Do(func() {
		globalLoader = &ConfigLoader{
			loadTimeout: 10 * time.Second, // Timeout for configuration loading
		}
	})
	return globalLoader
}

// SetLoadTimeout configures the timeout for configuration loading operations
func (cl *ConfigLoader) SetLoadTimeout(timeout time.Duration) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.loadTimeout = timeout
}

// LoadConfigSafe loads configuration with thread safety and caching
func (cl *ConfigLoader) LoadConfigSafe() (*museconfig.Config, error) {
	cl.mu.RLock()
	// Check if we have a valid cached config
	if cl.cachedConfig != nil && time.Now().Before(cl.cachedConfig.ExpiresAt) {
		config := cl.cachedConfig.Config
		cl.mu.RUnlock()
		return config, nil
	}
	cl.mu.RUnlock()

	// Need to reload - acquire write lock
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Double-check pattern - another goroutine might have loaded while we waited
	if cl.cachedConfig != nil && time.Now().Before(cl.cachedConfig.ExpiresAt) {
		return cl.cachedConfig.Config, nil
	}

	// Load configuration with timeout
	config, err := cl.loadConfigWithTimeout()
	if err != nil {
		return nil, err
	}

	// Cache the loaded configuration (5 minute expiration)
	cl.cachedConfig = &CachedConfig{
		Config:    config,
		LoadTime:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	return config, nil
}

// loadConfigWithTimeout loads configuration with a timeout to prevent hanging
func (cl *ConfigLoader) loadConfigWithTimeout() (*museconfig.Config, error) {
	done := make(chan struct {
		*museconfig.Config
		error
	}, 1)

	go func() {
		config, err := cl.loadConfigInternal()
		done <- struct {
			*museconfig.Config
			error
		}{config, err}
	}()

	select {
	case result := <-done:
		return result.Config, result.error
	case <-time.After(cl.loadTimeout):
		return nil, ConfigError{
			Stage:  "loading",
			Reason: fmt.Sprintf("timeout after %v", cl.loadTimeout),
		}
	}
}

// loadConfigInternal performs the actual configuration loading
func (cl *ConfigLoader) loadConfigInternal() (*museconfig.Config, error) {
	slog.Debug("Loading config with thread safety")
	k := koanf.New(".")

	// Atomically capture environment state
	env := cl.captureEnvironment()

	// Define the list of config file paths to check
	configPaths := cl.buildConfigPaths(env)

	// Load the first existing config file
	configPath, err := cl.loadConfigFile(k, configPaths)
	if err != nil {
		return nil, err
	}

	// Load environment variables safely
	if err := cl.loadEnvironmentVariables(k, env); err != nil {
		// Log but don't fail on environment variable errors
		slog.Warn("Failed to load environment variables", "error", err)
	}

	// Unmarshal into the struct
	var cfg museconfig.Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, ConfigError{
			Stage:  "unmarshaling",
			Path:   configPath,
			Reason: "failed to unmarshal configuration",
			Err:    err,
		}
	}

	// Safely handle API keys with environment fallback
	if err := cl.handleAPIKeys(&cfg, env); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Environment represents captured environment state
type Environment struct {
	Variables map[string]string
	Home      string
	XDGConfig string
}

// captureEnvironment atomically captures relevant environment variables
func (cl *ConfigLoader) captureEnvironment() Environment {
	return Environment{
		Variables: cl.captureEnvVars(),
		Home:      os.Getenv("HOME"),
		XDGConfig: os.Getenv("XDG_CONFIG_HOME"),
	}
}

// captureEnvVars captures all environment variables starting with MUSE_
func (cl *ConfigLoader) captureEnvVars() map[string]string {
	vars := make(map[string]string)

	// Get all environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		// Capture MUSE_ prefixed variables
		if strings.HasPrefix(key, "MUSE_") {
			vars[key] = value
		}

		// Capture API key variables
		if strings.HasSuffix(key, "_API_KEY") {
			vars[key] = value
		}
	}

	return vars
}

// buildConfigPaths builds the list of configuration file paths to check
func (cl *ConfigLoader) buildConfigPaths(env Environment) []string {
	paths := []string{
		"./muse.yaml", // Local directory
	}

	// Add XDG config path if available
	if env.XDGConfig != "" {
		paths = append(paths, env.XDGConfig+"/muse/muse.yaml")
	}

	// Add home config path if available
	if env.Home != "" {
		paths = append(paths, env.Home+"/.config/muse/muse.yaml")
	}

	return paths
}

// loadConfigFile loads the first available configuration file
func (cl *ConfigLoader) loadConfigFile(k *koanf.Koanf, configPaths []string) (string, error) {
	// Try to load from files
	for _, path := range configPaths {
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); err == nil {
			if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
				return "", ConfigError{
					Stage:  "file_loading",
					Path:   path,
					Reason: "failed to load config file",
					Err:    err,
				}
			}
			slog.Debug("Loaded config from file", "path", path)
			return path, nil
		}
	}

	// Fallback to example config
	if err := k.Load(rawbytes.Provider(museconfig.ExampleConfig), yaml.Parser()); err != nil {
		return "", ConfigError{
			Stage:  "example_loading",
			Reason: "failed to load example config",
			Err:    err,
		}
	}

	slog.Debug("Using example config")
	return "embedded_example", nil
}

// loadEnvironmentVariables loads environment variables safely
func (cl *ConfigLoader) loadEnvironmentVariables(k *koanf.Koanf, env Environment) error {
	// Create a custom environment provider using captured variables
	envProvider := &SafeEnvProvider{
		prefix:    "MUSE_",
		delimiter: ".",
		variables: env.Variables,
		transform: func(s string) string {
			return strings.ReplaceAll(strings.ToLower(s), "_", ".")
		},
	}

	if err := k.Load(envProvider, nil); err != nil {
		return ConfigError{
			Stage:  "env_loading",
			Reason: "failed to load environment variables",
			Err:    err,
		}
	}

	return nil
}

// handleAPIKeys safely handles API key configuration with environment fallback
func (cl *ConfigLoader) handleAPIKeys(cfg *museconfig.Config, env Environment) error {
	if cfg.LLM.Config == nil {
		cfg.LLM.Config = make(map[string]any)
	}

	// Create a copy of the config map to avoid concurrent modification
	configCopy := make(map[string]any)
	for k, v := range cfg.LLM.Config {
		configCopy[k] = v
	}

	// Handle API keys with environment fallback
	for key := range configCopy {
		envKey := strings.ToUpper(fmt.Sprintf("%s_API_KEY", key))
		envValue, exists := env.Variables[envKey]

		if exists && envValue != "" {
			// Validate the credential before using it
			if err := security.ValidateCredential(envValue); err != nil {
				slog.Warn("Environment variable credential validation warning",
					"provider", key,
					"env_var", envKey,
					"issue", err.Error(),
					"masked_value", security.MaskCredential(envValue))
			}

			// Use environment variable (already captured atomically)
			configCopy[key] = envValue

			slog.Debug("Using API key from environment",
				"provider", key,
				"env_var", envKey,
				"masked_value", security.MaskCredential(envValue))
		}
	}

	// Atomically replace the configuration
	cfg.LLM.Config = configCopy

	return nil
}

// SafeEnvProvider is a thread-safe environment variable provider
type SafeEnvProvider struct {
	prefix    string
	delimiter string
	variables map[string]string
	transform func(string) string
}

// ReadBytes implements the koanf Provider interface
func (p *SafeEnvProvider) ReadBytes() ([]byte, error) {
	return nil, fmt.Errorf("SafeEnvProvider does not support ReadBytes")
}

// Read implements the koanf Provider interface
func (p *SafeEnvProvider) Read() (map[string]interface{}, error) {
	out := make(map[string]interface{})

	for key, value := range p.variables {
		if !strings.HasPrefix(key, p.prefix) {
			continue
		}

		// Strip the prefix
		key = strings.TrimPrefix(key, p.prefix)

		// Transform the key if a transformer is provided
		if p.transform != nil {
			key = p.transform(key)
		}

		// Set the value
		out[key] = value
	}

	return out, nil
}
