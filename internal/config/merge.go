package config

// Overrides represents CLI overrides.
type Overrides struct {
	Backend string
	Model   string // Override model selection (fast/slow)
	Variant string
	Yes     bool
	Verbose bool
	Debug   bool
}

// EffectiveConfig represents the final configuration after merging defaults,
// config file, and CLI overrides.
type EffectiveConfig struct {
	Backend   string
	Command   string
	Args      []string
	ModelSlow string // Model for AGENTS.md and tasks.yaml generation
	ModelFast string // Model for implementation
	Model     string // Currently selected model (slow by default, or overridden)
	Variant   string
	Quiet     bool
	Yes       bool
	Verbose   bool
	Debug     bool
	Retry     Retry
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
		eff.ModelSlow = b.Models.Slow
		eff.ModelFast = b.Models.Fast
		eff.Model = b.Models.Slow // Default to slow model
		eff.Variant = b.Variant
	}

	// 3. Apply CLI Overrides
	if overrides.Model != "" {
		// Override can be "fast" or "slow" to switch between models, or a custom model name
		switch overrides.Model {
		case "fast":
			eff.Model = eff.ModelFast
		case "slow":
			eff.Model = eff.ModelSlow
		default:
			// Custom model name
			eff.Model = overrides.Model
		}
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
