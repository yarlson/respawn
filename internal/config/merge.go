package config

type Overrides struct {
	Backend string
	Model   string // "fast", "slow", or custom model name
	Variant string
	Yes     bool
	Verbose bool
	Debug   bool
}

type EffectiveConfig struct {
	Backend string
	Command string
	Args    []string
	Model   string
	Variant string
	Quiet   bool
	Yes     bool
	Verbose bool
	Debug   bool
	Retry   Retry
}

func Merge(cfg *Config, overrides Overrides) *EffectiveConfig {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	eff := &EffectiveConfig{
		Retry: cfg.Defaults.Retry,
		Quiet: cfg.Defaults.Quiet,
	}

	eff.Backend = cfg.Defaults.Backend
	if overrides.Backend != "" {
		eff.Backend = overrides.Backend
	}

	var modelFast, modelSlow string
	if b, ok := cfg.Backends[eff.Backend]; ok {
		eff.Command = b.Command
		eff.Args = b.Args
		eff.Variant = b.Variant
		modelFast = b.Models.Fast
		modelSlow = b.Models.Slow
	}

	switch overrides.Model {
	case "fast":
		eff.Model = modelFast
	case "slow":
		eff.Model = modelSlow
	case "":
		// No override - leave empty, command must specify
	default:
		eff.Model = overrides.Model
	}

	if overrides.Variant != "" {
		eff.Variant = overrides.Variant
	}
	if overrides.Yes {
		eff.Yes = true
	}
	if overrides.Verbose {
		eff.Verbose = true
		eff.Quiet = false
	}
	if overrides.Debug {
		eff.Debug = true
		eff.Verbose = true
		eff.Quiet = false
	}

	return eff
}
