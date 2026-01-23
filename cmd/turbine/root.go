package turbine

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/backends/claude"
	"github.com/yarlson/turbine/internal/backends/opencode"
	"github.com/yarlson/turbine/internal/config"
)

var (
	globalBackend string
	globalModel   string
	globalVariant string
	globalYes     bool
	globalVerbose bool
	globalDebug   bool
)

var rootCmd = &cobra.Command{
	Use:   "turbine",
	Short: "Spin up and execute the task manifest",
	Long:  `Reads tasks from .turbine/tasks.yaml and executes each one autonomously using an AI backend.`,
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalBackend, "backend", "", "AI backend (opencode, claude)")
	rootCmd.PersistentFlags().StringVar(&globalModel, "model", "", "Model name for the backend")
	rootCmd.PersistentFlags().StringVar(&globalVariant, "variant", "", "Variant configuration")
	rootCmd.PersistentFlags().BoolVar(&globalYes, "yes", false, "Skip confirmation prompts")
	rootCmd.PersistentFlags().BoolVar(&globalVerbose, "verbose", false, "Show detailed output")
	rootCmd.PersistentFlags().BoolVar(&globalDebug, "debug", false, "Show debug logs")
}

type namedBackend struct {
	backends.Backend
	name string
}

func (b *namedBackend) Name() string { return b.name }

func resolveBackend(cfg *config.Config, defaultModel string) (*namedBackend, string, error) {
	modelOverride := globalModel
	if modelOverride == "" {
		modelOverride = defaultModel
	}

	effCfg := config.Merge(cfg, config.Overrides{
		Backend: globalBackend,
		Model:   modelOverride,
		Variant: globalVariant,
	})

	bCfg, ok := cfg.Backends[effCfg.Backend]
	if !ok {
		return nil, "", fmt.Errorf("unknown backend: %s", effCfg.Backend)
	}

	var backend backends.Backend
	switch effCfg.Backend {
	case "opencode":
		backend = opencode.New(bCfg)
	case "claude":
		backend = claude.New(bCfg)
	default:
		return nil, "", fmt.Errorf("unsupported backend: %s", effCfg.Backend)
	}

	return &namedBackend{Backend: backend, name: effCfg.Backend}, effCfg.Model, nil
}

// resolveBackendWithModels returns a backend and both fast/slow model names.
// Used for two-phase operations like decompose that need both models.
func resolveBackendWithModels(cfg *config.Config) (*namedBackend, string, string, error) {
	backendName := globalBackend
	if backendName == "" {
		backendName = cfg.Defaults.Backend
	}

	bCfg, ok := cfg.Backends[backendName]
	if !ok {
		return nil, "", "", fmt.Errorf("unknown backend: %s", backendName)
	}

	// Apply variant override
	if globalVariant != "" {
		bCfg.Variant = globalVariant
	}

	var backend backends.Backend
	switch backendName {
	case "opencode":
		backend = opencode.New(bCfg)
	case "claude":
		backend = claude.New(bCfg)
	default:
		return nil, "", "", fmt.Errorf("unsupported backend: %s", backendName)
	}

	// If user specified a model, use it for both (override behavior)
	fastModel := bCfg.Models.Fast
	slowModel := bCfg.Models.Slow
	if globalModel != "" {
		fastModel = globalModel
		slowModel = globalModel
	}

	return &namedBackend{Backend: backend, name: backendName}, fastModel, slowModel, nil
}
