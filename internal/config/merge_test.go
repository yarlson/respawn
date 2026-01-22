package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	t.Run("defaults only", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{})

		assert.Equal(t, "opencode", eff.Backend)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.Model)
		assert.False(t, eff.Quiet)
		assert.False(t, eff.Verbose)
		assert.False(t, eff.Debug)
		assert.False(t, eff.Yes)
	})

	t.Run("config overrides defaults", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Defaults.Backend = "claude"
		cfg.Defaults.Quiet = true

		eff := Merge(cfg, Overrides{})

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.Model)
		assert.True(t, eff.Quiet)
	})

	t.Run("CLI overrides config", func(t *testing.T) {
		cfg := DefaultConfig()
		overrides := Overrides{
			Backend: "claude",
			Model:   "gpt-4",
			Verbose: true,
			Yes:     true,
		}

		eff := Merge(cfg, overrides)

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "gpt-4", eff.Model)
		assert.True(t, eff.Verbose)
		assert.False(t, eff.Quiet) // Verbose should turn off Quiet
		assert.True(t, eff.Yes)
	})

	t.Run("Debug implies Verbose and Quiet false", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Defaults.Quiet = true
		overrides := Overrides{
			Debug: true,
		}

		eff := Merge(cfg, overrides)

		assert.True(t, eff.Debug)
		assert.True(t, eff.Verbose)
		assert.False(t, eff.Quiet)
	})

	t.Run("Variant override", func(t *testing.T) {
		cfg := DefaultConfig()
		overrides := Overrides{
			Variant: "experimental",
		}

		eff := Merge(cfg, overrides)
		assert.Equal(t, "experimental", eff.Variant)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		eff := Merge(nil, Overrides{Backend: "claude"})
		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "claude", eff.Command)
	})

	t.Run("non-existent backend in overrides", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{Backend: "ghost"})

		assert.Equal(t, "ghost", eff.Backend)
		assert.Equal(t, "", eff.Command) // Backend not found in config
	})
}
