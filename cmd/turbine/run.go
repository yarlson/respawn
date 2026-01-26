package turbine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/run"
	"github.com/yarlson/turbine/internal/ui"
)

var runPrdPath string

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := gitx.RepoRoot(ctx, cwd)
	if err != nil {
		return err
	}

	prdDestPath := filepath.Join(repoRoot, run.PRDRelPath)
	if runPrdPath != "" {
		if _, err := os.Stat(runPrdPath); os.IsNotExist(err) {
			return fmt.Errorf("PRD file not found: %s", runPrdPath)
		} else if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Join(repoRoot, ".turbine"), 0755); err != nil {
			return err
		}

		if _, err := os.Stat(prdDestPath); err == nil && !globalYes {
			fmt.Printf("%s exists. %s [y/N]: ", ui.Dim(prdDestPath), ui.Yellow("Overwrite?"))
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			resp := scanner.Text()
			if strings.ToLower(resp) != "y" {
				return fmt.Errorf("canceled")
			}
		}

		prdContent, err := os.ReadFile(runPrdPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(prdDestPath, prdContent, 0644); err != nil {
			return err
		}
	} else {
		if _, err := os.Stat(prdDestPath); os.IsNotExist(err) {
			return fmt.Errorf("PRD file not found: %s (use --prd)", prdDestPath)
		} else if err != nil {
			return err
		}
	}

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

	backend, fastModel, slowModel, err := resolveBackendWithModels(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Using backend: %s, fast: %s, slow: %s\n", backend.Name(), fastModel.Name, slowModel.Name)

	if r.Resume {
		fmt.Printf("Continuing from checkpoint: %s\n", r.State.RunID)
	}

	return r.Run(ctx, backend, run.Models{Fast: fastModel, Slow: slowModel})
}

func init() {
	rootCmd.RunE = runCmd
	rootCmd.Flags().StringVar(&runPrdPath, "prd", "", "Path to the PRD file")
}
