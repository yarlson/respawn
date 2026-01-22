package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"respawn/internal/backends/claude"
	"respawn/internal/backends/opencode"
	"respawn/internal/config"
	"respawn/internal/decomposer"
	"respawn/internal/gitx"
	"respawn/internal/run"

	"github.com/spf13/cobra"
)

var (
	prdPath string
)

var decomposeCmd = &cobra.Command{
	Use:   "decompose",
	Short: "Decompose a PRD into tasks",
	Long:  `Decompose takes a PRD file and breaks it down into actionable tasks in .respawn/tasks.yaml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDecompose(cmd)
	},
}

func runDecompose(cmd *cobra.Command) error {
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
			fmt.Printf("Missing required .gitignore entries: %v\n", missingIgnores)
			fmt.Print("Add them now? [y/N]: ")
			var resp string
			_, _ = fmt.Scanln(&resp)
			if strings.ToLower(resp) == "y" {
				if err := gitx.AddIgnoresToGitignore(repoRoot, missingIgnores); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("required .gitignore entries missing")
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
			fmt.Printf("%s already exists. Overwrite? [y/N]: ", tasksPath)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			resp := scanner.Text()
			if strings.ToLower(resp) != "y" {
				return fmt.Errorf("aborted by user")
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
		return fmt.Errorf("backend %s not fully supported for decomposition yet", backendName)
	}

	// 5. Initialize Artifacts for logging
	runID := run.GenerateRunID()
	artifacts, err := run.NewArtifacts(repoRoot, runID)
	if err != nil {
		return err
	}

	d := decomposer.New(backend, repoRoot)
	fmt.Printf("Decomposing PRD: %s using %s backend...\n", prdPath, backendName)
	if err := d.Decompose(ctx, prdPath, artifacts.Root()); err != nil {
		return err
	}

	fmt.Printf("Successfully decomposed PRD into %s (Run ID: %s)\n", tasksPath, runID)
	return nil
}

func init() {
	rootCmd.AddCommand(decomposeCmd)

	decomposeCmd.Flags().StringVar(&prdPath, "prd", "", "Path to the PRD file (required)")
	_ = decomposeCmd.MarkFlagRequired("prd")
}
