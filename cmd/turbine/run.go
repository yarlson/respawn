package turbine

import (
	"fmt"

	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/backends/claude"
	"github.com/yarlson/turbine/internal/backends/opencode"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/run"

	"github.com/spf13/cobra"
)

// runCmd is the action for the root command when invoked without subcommands.
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

	backendName := globalBackend
	if backendName == "" {
		backendName = cfg.Defaults.Backend
	}

	bCfg, ok := cfg.Backends[backendName]
	if !ok {
		return fmt.Errorf("unknown backend: %s", backendName)
	}

	// Apply CLI overrides
	if globalModel != "" {
		bCfg.Model = globalModel
	}
	if globalVariant != "" {
		bCfg.Variant = globalVariant
	}

	var backend backends.Backend
	switch backendName {
	case "opencode":
		backend = opencode.New(bCfg)
	case "claude":
		backend = claude.New(bCfg.Command, bCfg.Args)
	default:
		return fmt.Errorf("unsupported backend: %s", backendName)
	}

	r.PrintSummary()

	if r.Resume {
		fmt.Printf("Continuing from checkpoint: %s\n", r.State.RunID)
	}

	return r.Run(ctx, backend)
}

func init() {
	rootCmd.RunE = runCmd
}
