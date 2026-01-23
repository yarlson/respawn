package respawn

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yarlson/respawn/internal/backends/claude"
	"github.com/yarlson/respawn/internal/backends/opencode"
	"github.com/yarlson/respawn/internal/config"
	"github.com/yarlson/respawn/internal/decomposer"
	"github.com/yarlson/respawn/internal/gitx"
	"github.com/yarlson/respawn/internal/run"
	"github.com/yarlson/respawn/internal/ui"

	"github.com/spf13/cobra"
)

var (
	prdPath string
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load a PRD into the mission file",
	Long:  `Reads a PRD file and generates .respawn/tasks.yaml with executable tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLoad(cmd)
	},
}

func runLoad(cmd *cobra.Command) error {
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

	missingIgnores, err := gitx.MissingRespawnIgnores(ctx, repoRoot)
	if err != nil {
		return err
	}
	if len(missingIgnores) > 0 {
		if globalYes {
			if err := gitx.AddIgnoresToGitignore(repoRoot, missingIgnores); err != nil {
				return err
			}
		} else {
			fmt.Printf("%s\n", ui.Section("⚠", "These paths need to be in .gitignore"))
			for _, path := range missingIgnores {
				fmt.Printf("  %s\n", ui.Dim(path))
			}
			fmt.Print(ui.Yellow("Add them? [y/N]: "))
			var resp string
			_, _ = fmt.Scanln(&resp)
			if strings.ToLower(resp) == "y" {
				if err := gitx.AddIgnoresToGitignore(repoRoot, missingIgnores); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("canceled: .gitignore entries required")
			}
		}
	}

	// 2. Read PRD file existence check
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return fmt.Errorf("PRD file not found: %s", prdPath)
	}

	// 3. Check if tasks.yaml exists
	tasksPath := filepath.Join(repoRoot, ".respawn", "tasks.yaml")
	if _, err := os.Stat(tasksPath); err == nil {
		if !globalYes {
			fmt.Printf("%s exists. %s [y/N]: ", ui.Dim(tasksPath), ui.Yellow("Overwrite?"))
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

	// Apply CLI overrides to backend config
	if globalModel != "" {
		bCfg.Model = globalModel
	}
	if globalVariant != "" {
		bCfg.Variant = globalVariant
	}

	var backend decomposer.Backend
	switch backendName {
	case "opencode":
		backend = opencode.New(bCfg)
	case "claude":
		backend = claude.New(bCfg.Command, bCfg.Args)
	default:
		return fmt.Errorf("backend %s does not support decompose", backendName)
	}

	// 5. Initialize Artifacts for logging
	runID := run.GenerateRunID()
	artifacts, err := run.NewArtifacts(repoRoot, runID)
	if err != nil {
		return err
	}

	d := decomposer.New(backend, repoRoot)
	spinner := ui.NewSpinner(fmt.Sprintf("Generating tasks from %s (%s)...", prdPath, backendName))
	cancel := spinner.Start(ctx)
	defer cancel()

	if err := d.Decompose(ctx, prdPath, artifacts.Root()); err != nil {
		spinner.Fail(fmt.Sprintf("Failed: %v", err))
		return err
	}

	spinner.Stop(fmt.Sprintf("Created %s (run %s)", tasksPath, runID))
	return nil
}

func init() {
	rootCmd.AddCommand(loadCmd)

	loadCmd.Flags().StringVar(&prdPath, "prd", "", "Path to the PRD file")
	_ = loadCmd.MarkFlagRequired("prd")
}
