package turbine

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/run"
)

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	r, err := run.NewRunner(ctx, run.Config{
		AutoAddIgnore: globalYes,
		Defaults:      cfg.Defaults,
	})
	if err != nil {
		return err
	}

	backend, model, variant, err := resolveBackend(cfg, "fast")
	if err != nil {
		return err
	}

	r.PrintSummary()
	fmt.Printf("Using backend: %s, model: %s\n", backend.Name(), model)

	if r.Resume {
		fmt.Printf("Continuing from checkpoint: %s\n", r.State.RunID)
	}

	return r.Run(ctx, backend, model, variant)
}

func init() {
	rootCmd.RunE = runCmd
}
