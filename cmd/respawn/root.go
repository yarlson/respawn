package respawn

import (
	"github.com/spf13/cobra"
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
	Use:   "respawn",
	Short: "Run autonomous coding tasks from a task file",
	Long:  `Reads tasks from .respawn/tasks.yaml and executes each one autonomously using an AI backend.`,
}

// RootCmd returns the root cobra command.
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
