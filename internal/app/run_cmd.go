package app

import (
	"fmt"

	"respawn/internal/run"

	"github.com/spf13/cobra"
)

// runCmd is the action for the root command when invoked without subcommands.
func runCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	r, err := run.NewRunner(ctx, run.Config{
		AutoAddIgnore: globalYes,
	})
	if err != nil {
		return err
	}

	r.PrintSummary()

	if r.Resume {
		fmt.Printf("Resuming run %s\n", r.State.RunID)
	}

	// Task execution logic will be implemented in future tasks.
	return nil
}

func init() {
	rootCmd.RunE = runCmd
}
