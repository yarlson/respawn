package turbine

import (
	"github.com/spf13/cobra"
	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/config"
	relayprovider "github.com/yarlson/turbine/internal/relay/provider"
)

var (
	globalBackend string
	globalModel   string
	globalVariant string
	globalYes     bool
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
}

func resolveBackend(cfg *config.Config, defaultModel string) (relay.Provider, string, string, error) {
	return relayprovider.Resolve(cfg, config.Overrides{
		Backend: globalBackend,
		Model:   globalModel,
		Variant: globalVariant,
	}, defaultModel)
}

// resolveBackendWithModels returns a backend and both fast/slow model configs.
// Used for two-phase operations like decompose that need both models.
func resolveBackendWithModels(cfg *config.Config) (relay.Provider, config.Model, config.Model, error) {
	return relayprovider.ResolveWithModels(cfg, config.Overrides{
		Backend: globalBackend,
		Model:   globalModel,
		Variant: globalVariant,
	})
}
