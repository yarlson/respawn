package config

// Overrides represents CLI overrides.
type Overrides struct {
	Backend string
	Model   string
	Variant string
	Yes     bool
	Verbose bool
	Debug   bool
}

// EffectiveConfig represents the final configuration after merging defaults,
// config file, and CLI overrides.
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

// Merge computes the effective settings from the base config and CLI overrides.
// Precedence: CLI flags (when non-empty / true) override config; config overrides defaults.
func Merge(cfg *Config, overrides Overrides) *EffectiveConfig {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	eff := &EffectiveConfig{
		Retry: cfg.Defaults.Retry,
		Quiet: cfg.Defaults.Quiet,
	}

	// 1. Determine Backend
	eff.Backend = cfg.Defaults.Backend
	if overrides.Backend != "" {
		eff.Backend = overrides.Backend
	}

	// 2. Load Backend settings from config
	if b, ok := cfg.Backends[eff.Backend]; ok {
		eff.Command = b.Command
		eff.Args = b.Args
		eff.Model = b.Model
		eff.Variant = b.Variant
	}

	// 3. Apply CLI Overrides
	if overrides.Model != "" {
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
