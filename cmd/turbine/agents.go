package turbine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yarlson/turbine/internal/agents"
	"github.com/yarlson/turbine/internal/backends/claude"
	"github.com/yarlson/turbine/internal/backends/opencode"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/run"
	"github.com/yarlson/turbine/internal/ui"

	"github.com/spf13/cobra"
)

var (
	agentsPrdPath string
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Generate AGENTS.md from a PRD with progressive disclosure",
	Long:  `Reads a PRD file and generates AGENTS.md with related documentation following progressive disclosure principles. Creates a CLAUDE.md symlink to AGENTS.md.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAgents(cmd)
	},
}

func runAgents(cmd *cobra.Command) error {
	ctx := cmd.Context()

	// 1. Preflight
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := gitx.RepoRoot(ctx, cwd)
	if err != nil {
		return err
	}

	// 2. Read PRD file existence check
	if _, err := os.Stat(agentsPrdPath); os.IsNotExist(err) {
		return fmt.Errorf("PRD file not found: %s", agentsPrdPath)
	}

	// 3. Check if AGENTS.md exists
	agentsMdPath := filepath.Join(repoRoot, "AGENTS.md")
	if _, err := os.Stat(agentsMdPath); err == nil {
		if !globalYes {
			fmt.Printf("%s exists. %s [y/N]: ", ui.Dim(agentsMdPath), ui.Yellow("Overwrite?"))
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			resp := scanner.Text()
			if strings.ToLower(resp) != "y" {
				return fmt.Errorf("canceled")
			}
		}
	}

	// 4. Load config and select backend
	cfg, err := config.Load()
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

	// Apply CLI overrides to backend config (variant only, model is slow by default for agents generation)
	if globalVariant != "" {
		bCfg.Variant = globalVariant
	}

	var backend agents.Backend
	switch backendName {
	case "opencode":
		backend = opencode.New(bCfg)
	case "claude":
		backend = claude.New(bCfg.Command, bCfg.Args)
	default:
		return fmt.Errorf("backend %s does not support agents generation", backendName)
	}

	// 5. Initialize Artifacts for logging
	runID := run.GenerateRunID()
	artifacts, err := run.NewArtifacts(repoRoot, runID)
	if err != nil {
		return err
	}

	generator := agents.New(backend, repoRoot)
	spinner := ui.NewSpinner(fmt.Sprintf("Generating AGENTS.md from %s (%s)...", agentsPrdPath, backendName))
	cancel := spinner.Start(ctx)
	defer cancel()

	if err := generator.Generate(ctx, agentsPrdPath, artifacts.Root()); err != nil {
		spinner.Fail(fmt.Sprintf("Failed: %v", err))
		return err
	}

	spinner.Stop(fmt.Sprintf("Created AGENTS.md and supporting documentation (run %s)", runID))
	return nil
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	agentsCmd.Flags().StringVar(&agentsPrdPath, "prd", "", "Path to the PRD file")
	_ = agentsCmd.MarkFlagRequired("prd")
}
