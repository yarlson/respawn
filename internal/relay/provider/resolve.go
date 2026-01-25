package relayprovider

import (
	"fmt"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/relay/provider"
	"github.com/yarlson/relay/provider/claude"
	"github.com/yarlson/relay/provider/opencode"
	"github.com/yarlson/turbine/internal/config"
)

// Resolve returns a relay provider and effective model/variant.
func Resolve(cfg *config.Config, overrides config.Overrides, defaultModel string) (relay.Provider, string, string, error) {
	modelOverride := overrides.Model
	if modelOverride == "" {
		modelOverride = defaultModel
	}

	effCfg := config.Merge(cfg, config.Overrides{
		Backend: overrides.Backend,
		Model:   modelOverride,
		Variant: overrides.Variant,
	})

	bCfg, ok := cfg.Backends[effCfg.Backend]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown backend: %s", effCfg.Backend)
	}

	provCfg := provider.Config{
		Executable: bCfg.Command,
		Args:       bCfg.Args,
	}

	var p relay.Provider
	var err error
	switch effCfg.Backend {
	case "opencode":
		p, err = opencode.New(provCfg)
	case "claude":
		p, err = claude.New(provCfg)
	default:
		return nil, "", "", fmt.Errorf("unsupported backend: %s", effCfg.Backend)
	}
	if err != nil {
		return nil, "", "", err
	}

	return p, effCfg.Model, effCfg.Variant, nil
}

// ResolveWithModels returns a relay provider and both fast/slow model configs.
func ResolveWithModels(cfg *config.Config, overrides config.Overrides) (relay.Provider, config.Model, config.Model, error) {
	backendName := overrides.Backend
	if backendName == "" {
		backendName = cfg.Defaults.Backend
	}

	bCfg, ok := cfg.Backends[backendName]
	if !ok {
		return nil, config.Model{}, config.Model{}, fmt.Errorf("unknown backend: %s", backendName)
	}

	provCfg := provider.Config{
		Executable: bCfg.Command,
		Args:       bCfg.Args,
	}

	var p relay.Provider
	var err error
	switch backendName {
	case "opencode":
		p, err = opencode.New(provCfg)
	case "claude":
		p, err = claude.New(provCfg)
	default:
		return nil, config.Model{}, config.Model{}, fmt.Errorf("unsupported backend: %s", backendName)
	}
	if err != nil {
		return nil, config.Model{}, config.Model{}, err
	}

	fast := bCfg.Models.Fast
	slow := bCfg.Models.Slow

	if overrides.Model != "" {
		fast.Name = overrides.Model
		slow.Name = overrides.Model
	}
	if overrides.Variant != "" {
		fast.Variant = overrides.Variant
		slow.Variant = overrides.Variant
	}

	return p, fast, slow, nil
}
