package turbine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yarlson/turbine/internal/agents"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/run"
	"github.com/yarlson/turbine/internal/ui"
)

var agentsPrdPath string

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Generate AGENTS.md from a PRD",
	RunE:  runAgents,
}

func runAgents(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := gitx.RepoRoot(ctx, cwd)
	if err != nil {
		return err
	}

	if _, err := os.Stat(agentsPrdPath); os.IsNotExist(err) {
		return fmt.Errorf("PRD file not found: %s", agentsPrdPath)
	}

	agentsMdPath := filepath.Join(repoRoot, "AGENTS.md")
	if _, err := os.Stat(agentsMdPath); err == nil && !globalYes {
		fmt.Printf("%s exists. %s [y/N]: ", ui.Dim(agentsMdPath), ui.Yellow("Overwrite?"))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.ToLower(scanner.Text()) != "y" {
			return fmt.Errorf("canceled")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	backend, model, variant, err := resolveBackend(cfg, "slow")
	if err != nil {
		return err
	}

	runID := run.GenerateRunID()
	artifacts, err := run.NewArtifacts(repoRoot, runID)
	if err != nil {
		return err
	}

	generator := agents.New(backend, repoRoot)
	spinner := ui.NewSpinner(fmt.Sprintf("Generating AGENTS.md from %s (%s, %s)...", agentsPrdPath, backend.Name(), model))
	cancel := spinner.Start(ctx)
	defer cancel()

	if err := generator.Generate(ctx, agentsPrdPath, artifacts.Root(), model, variant); err != nil {
		spinner.Fail(fmt.Sprintf("Failed: %v", err))
		return err
	}

	spinner.Stop(fmt.Sprintf("Created AGENTS.md (run %s)", runID))
	return nil
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.Flags().StringVar(&agentsPrdPath, "prd", "", "Path to the PRD file")
	_ = agentsCmd.MarkFlagRequired("prd")
}
