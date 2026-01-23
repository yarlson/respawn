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
		assert.Equal(t, "anthropic/claude-opus-4-5", eff.Model) // Slow model by default
		assert.Equal(t, "anthropic/claude-opus-4-5", eff.ModelSlow)
		assert.Equal(t, "anthropic/claude-sonnet", eff.ModelFast)
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
		assert.Equal(t, "claude-4-5-opus-latest", eff.Model) // Slow model by default
		assert.Equal(t, "claude-4-5-opus-latest", eff.ModelSlow)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.ModelFast)
		assert.True(t, eff.Quiet)
	})

	t.Run("CLI overrides config with custom model", func(t *testing.T) {
		cfg := DefaultConfig()
		overrides := Overrides{
			Backend: "claude",
			Model:   "gpt-4",
			Verbose: true,
			Yes:     true,
		}

		eff := Merge(cfg, overrides)

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "gpt-4", eff.Model) // Custom model
		assert.True(t, eff.Verbose)
		assert.False(t, eff.Quiet) // Verbose should turn off Quiet
		assert.True(t, eff.Yes)
	})

	t.Run("CLI switches to fast model", func(t *testing.T) {
		cfg := DefaultConfig()
		overrides := Overrides{
			Backend: "claude",
			Model:   "fast",
		}

		eff := Merge(cfg, overrides)

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.Model)
		assert.Equal(t, "claude-4-5-opus-latest", eff.ModelSlow)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.ModelFast)
	})

	t.Run("CLI switches to slow model explicitly", func(t *testing.T) {
		cfg := DefaultConfig()
		overrides := Overrides{
			Backend: "claude",
			Model:   "slow",
		}

		eff := Merge(cfg, overrides)

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "claude-4-5-opus-latest", eff.Model)
		assert.Equal(t, "claude-4-5-opus-latest", eff.ModelSlow)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.ModelFast)
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
		assert.Equal(t, "claude-4-5-opus-latest", eff.Model) // Slow model by default
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.ModelFast)
	})

	t.Run("non-existent backend in overrides", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{Backend: "ghost"})

		assert.Equal(t, "ghost", eff.Backend)
		assert.Equal(t, "", eff.Command)   // Backend not found in config
		assert.Equal(t, "", eff.ModelSlow) // No models configured
		assert.Equal(t, "", eff.ModelFast)
	})
}
