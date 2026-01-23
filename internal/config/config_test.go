package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveConfigPath(t *testing.T) {
	t.Run("XDG_CONFIG_HOME set", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", oldXDG) }()

		tmpDir := t.TempDir()
		err := os.Setenv("XDG_CONFIG_HOME", tmpDir)
		require.NoError(t, err)

		expected := filepath.Join(tmpDir, "turbine", "turbine.yaml")
		assert.Equal(t, expected, ResolveConfigPath())
	})

	t.Run("XDG_CONFIG_HOME unset", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", oldXDG) }()

		err := os.Unsetenv("XDG_CONFIG_HOME")
		require.NoError(t, err)

		home, err := os.UserHomeDir()
		require.NoError(t, err)

		expected := filepath.Join(home, ".config", "turbine", "turbine.yaml")
		assert.Equal(t, expected, ResolveConfigPath())
	})
}

func TestLoad(t *testing.T) {
	t.Run("missing file returns defaults", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", oldXDG) }()

		// Point XDG_CONFIG_HOME to a non-existent path
		err := os.Setenv("XDG_CONFIG_HOME", "/non-existent-path-12345")
		require.NoError(t, err)

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, DefaultConfig(), cfg)
	})

	t.Run("valid YAML overrides defaults", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", oldXDG) }()

		tmpDir := t.TempDir()
		err := os.Setenv("XDG_CONFIG_HOME", tmpDir)
		require.NoError(t, err)

		configDir := filepath.Join(tmpDir, "turbine")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		yamlData := `
defaults:
  backend: "claude"
  quiet: true
  retry:
    cycles: 5
    attempts: 10
backends:
  opencode:
    command: "custom-opencode"
    args: ["--debug"]
    model: "claude-3-opus"
    variant: "fast"
`
		err = os.WriteFile(filepath.Join(configDir, "turbine.yaml"), []byte(yamlData), 0644)
		require.NoError(t, err)

		cfg, err := Load()
		assert.NoError(t, err)

		assert.Equal(t, "claude", cfg.Defaults.Backend)
		assert.True(t, cfg.Defaults.Quiet)
		assert.Equal(t, 5, cfg.Defaults.Retry.Cycles)
		assert.Equal(t, 10, cfg.Defaults.Retry.Attempts)

		assert.Equal(t, "custom-opencode", cfg.Backends["opencode"].Command)
		assert.Equal(t, []string{"--debug"}, cfg.Backends["opencode"].Args)
		assert.Equal(t, "claude-3-opus", cfg.Backends["opencode"].Model)
		assert.Equal(t, "fast", cfg.Backends["opencode"].Variant)

		// Ensure claude backend still has defaults if not overridden
		assert.Equal(t, "claude", cfg.Backends["claude"].Command)
	})

	t.Run("invalid YAML returns error", func(t *testing.T) {
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		defer func() { _ = os.Setenv("XDG_CONFIG_HOME", oldXDG) }()

		tmpDir := t.TempDir()
		err := os.Setenv("XDG_CONFIG_HOME", tmpDir)
		require.NoError(t, err)

		configDir := filepath.Join(tmpDir, "turbine")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(configDir, "turbine.yaml"), []byte("invalid: yaml: :"), 0644)
		require.NoError(t, err)

		cfg, err := Load()
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}
