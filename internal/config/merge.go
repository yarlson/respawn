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

	var fast, slow Model
	if b, ok := cfg.Backends[eff.Backend]; ok {
		eff.Command = b.Command
		eff.Args = b.Args
		fast = b.Models.Fast
		slow = b.Models.Slow
	}

	switch overrides.Model {
	case "fast":
		eff.Model = fast.Name
		eff.Variant = fast.Variant
	case "slow":
		eff.Model = slow.Name
		eff.Variant = slow.Variant
	case "":
		// No override - leave empty, command must specify
	default:
		eff.Model = overrides.Model
	}

	// CLI variant override takes precedence
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
