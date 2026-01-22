package app

import (
	"github.com/spf13/cobra"
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
