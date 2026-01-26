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
	Rotations int `yaml:"rotations"`
	Strokes   int `yaml:"strokes"`
}

// Model holds a model name and optional variant.
type Model struct {
	Name    string `yaml:"name"`
	Variant string `yaml:"variant,omitempty"`
}

// UnmarshalYAML supports both string and struct format for backward compatibility.
// String format: "model-name"
// Struct format: {name: "model-name", variant: "high"}
func (m *Model) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try string first (backward compat)
	var s string
	if err := unmarshal(&s); err == nil {
		m.Name = s
		return nil
	}

	// Try struct
	type modelAlias Model
	var alias modelAlias
	if err := unmarshal(&alias); err != nil {
		return err
	}
	*m = Model(alias)
	return nil
}

// Models holds model configurations for different use cases.
type Models struct {
	Fast Model `yaml:"fast"` // Used for implementation tasks
	Slow Model `yaml:"slow"` // Used for AGENTS.md and task planning
}

// Backend holds configuration for a specific agent backend.
type Backend struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Models  Models   `yaml:"models"`
}

// DefaultConfig returns a Config with hardcoded defaults.
func DefaultConfig() *Config {
	return &Config{
		Defaults: Defaults{
			Backend: "opencode",
			Quiet:   false,
			Retry: Retry{
				Rotations: 3,
				Strokes:   3,
			},
		},
		Backends: map[string]Backend{
			"opencode": {
				Command: "opencode",
				Args:    []string{},
				Models: Models{
					Slow: Model{Name: "anthropic/claude-opus-4-5"},
					Fast: Model{Name: "anthropic/claude-sonnet"},
				},
			},
			"claude": {
				Command: "claude",
				Args:    []string{"-p", "--dangerously-skip-permissions"},
				Models: Models{
					Slow: Model{Name: "claude-4-5-opus-latest"},
					Fast: Model{Name: "claude-3-5-sonnet-latest"},
				},
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
