package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	t.Run("no model override", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{})

		assert.Equal(t, "opencode", eff.Backend)
		assert.Equal(t, "", eff.Model)
		assert.False(t, eff.Quiet)
	})

	t.Run("fast model", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{Backend: "claude", Model: "fast"})

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "claude-3-5-sonnet-latest", eff.Model)
	})

	t.Run("slow model", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{Backend: "claude", Model: "slow"})

		assert.Equal(t, "claude", eff.Backend)
		assert.Equal(t, "claude-4-5-opus-latest", eff.Model)
	})

	t.Run("custom model", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{Model: "gpt-4"})

		assert.Equal(t, "gpt-4", eff.Model)
	})

	t.Run("debug implies verbose", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Defaults.Quiet = true
		eff := Merge(cfg, Overrides{Debug: true})

		assert.True(t, eff.Debug)
		assert.True(t, eff.Verbose)
		assert.False(t, eff.Quiet)
	})

	t.Run("unknown backend", func(t *testing.T) {
		cfg := DefaultConfig()
		eff := Merge(cfg, Overrides{Backend: "ghost"})

		assert.Equal(t, "ghost", eff.Backend)
		assert.Equal(t, "", eff.Command)
	})
}
