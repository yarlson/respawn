package app

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
	Short: "Respawn is a minimal, reliable task execution harness",
	Long:  `Respawn reads tasks from .respawn/tasks.yaml and executes them autonomously.`,
}

// RootCmd returns the root cobra command.
func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().StringVar(&globalBackend, "backend", "", "Backend to use")
	rootCmd.PersistentFlags().StringVar(&globalModel, "model", "", "Model to use")
	rootCmd.PersistentFlags().StringVar(&globalVariant, "variant", "", "Variant to use")
	rootCmd.PersistentFlags().BoolVar(&globalYes, "yes", false, "Auto-approve prompts")
	rootCmd.PersistentFlags().BoolVar(&globalVerbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&globalDebug, "debug", false, "Enable debug logging")
}
