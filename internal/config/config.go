package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the global configuration for turbine.
type Config struct {
	Defaults Defaults           `yaml:"defaults"`
	Backends map[string]Backend `yaml:"backends"`
}

// Defaults holds default settings for turbine.
type Defaults struct {
	Backend string `yaml:"backend"`
	Quiet   bool   `yaml:"quiet"`
	Retry   Retry  `yaml:"retry"`
}

// Retry holds retry configuration.
type Retry struct {
	Cycles   int `yaml:"cycles"`
	Attempts int `yaml:"attempts"`
}

// Backend holds configuration for a specific agent backend.
type Backend struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Model   string   `yaml:"model"`
	Variant string   `yaml:"variant,omitempty"`
}

// DefaultConfig returns a Config with hardcoded defaults.
func DefaultConfig() *Config {
	return &Config{
		Defaults: Defaults{
			Backend: "opencode",
			Quiet:   false,
			Retry: Retry{
				Cycles:   3,
				Attempts: 3,
			},
		},
		Backends: map[string]Backend{
			"opencode": {
				Command: "opencode",
				Args:    []string{},
				Model:   "anthropic/claude-opus-4-5",
			},
			"claude": {
				Command: "claude",
				Args:    []string{"-p", "--dangerously-skip-permissions"},
				Model:   "claude-4-5-opus-latest",
			},
		},
	}
}

// ResolveConfigPath returns the absolute path to the global config file.
func ResolveConfigPath() string {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "turbine", "turbine.yaml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "turbine", "turbine.yaml")
}

// Load loads the configuration from the global config path.
// If the file is missing, it returns the default configuration without error.
func Load() (*Config, error) {
	path := ResolveConfigPath()
	if path == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
