package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd is the action for the root command when invoked without subcommands.
func runCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("Running tasks from .respawn/tasks.yaml...")
	// Placeholder for task execution logic.
	return nil
}

func init() {
	rootCmd.RunE = runCmd
}
