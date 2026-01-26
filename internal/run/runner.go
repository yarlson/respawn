package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/decomposer"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/state"
	"github.com/yarlson/turbine/internal/tasks"
	"github.com/yarlson/turbine/internal/ui"
)

type Runner struct {
	RepoRoot     string
	TaskFile     *tasks.TaskFile
	State        *state.RunState
	Config       config.Defaults
	Resume       bool
	PRDPath      string
	ProgressPath string
}

type Config struct {
	AutoAddIgnore bool
	Cwd           string
	Defaults      config.Defaults
}

type Models struct {
	Fast config.Model
	Slow config.Model
}

func NewRunner(ctx context.Context, cfg Config) (*Runner, error) {
	cwd := cfg.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get current working directory: %w", err)
		}
	}

	repoRoot, err := gitx.RepoRoot(ctx, cwd)
	if err != nil {
		return nil, fmt.Errorf("determine repo root: %w", err)
	}

	prdPath := filepath.Join(repoRoot, PRDRelPath)
	if _, err := os.Stat(prdPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PRD file not found: %s", prdPath)
	} else if err != nil {
		return nil, fmt.Errorf("stat PRD file: %w", err)
	}

	progressPath := filepath.Join(repoRoot, ProgressRelPath)

	// Resume state
	runState, exists, err := state.Load(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	if !exists {
		runState = &state.RunState{
			RunID: GenerateRunID(),
		}
	}

	// Clean working tree rule
	// IMPORTANT: Check this BEFORE we modify .gitignore ourselves
	if !exists {
		dirty, err := gitx.IsDirty(ctx, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("check repository status: %w", err)
		}
		if dirty {
			return nil, fmt.Errorf("uncommitted changes detected; save your progress before starting")
		}
	}

	// .gitignore check
	missing, err := gitx.MissingTurbineIgnores(ctx, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("check .gitignore: %w", err)
	}

	if len(missing) > 0 {
		if cfg.AutoAddIgnore {
			if err := gitx.AddIgnoresToGitignore(repoRoot, missing); err != nil {
				return nil, fmt.Errorf("update .gitignore: %w", err)
			}
		} else {
			fmt.Printf("%s\n", ui.Section("⚠", "Not in .gitignore"))
			for _, m := range missing {
				fmt.Printf("  %s\n", ui.Dim(m))
			}
			fmt.Printf("%s\n", ui.Dim("Use --yes to add automatically."))
		}
	}

	if _, err := EnsureProgressFile(repoRoot); err != nil {
		return nil, err
	}

	return &Runner{
		RepoRoot:     repoRoot,
		State:        runState,
		Config:       cfg.Defaults,
		Resume:       exists,
		PRDPath:      prdPath,
		ProgressPath: progressPath,
	}, nil
}

func (r *Runner) PrintSummary() {
	if r.TaskFile == nil {
		fmt.Printf("\n%s\n", ui.Divider(40))
		fmt.Printf("%s\n", ui.Dim("No active task loaded."))
		return
	}

	fmt.Printf("\n%s\n", ui.Divider(40))
	fmt.Printf("task: %s %s (%s)\n", r.TaskFile.Task.ID, r.TaskFile.Task.Title, r.TaskFile.Task.Status)
}

// Run plans and executes tasks sequentially using the provided backend and models.
func (r *Runner) Run(ctx context.Context, backend relay.Provider, models Models) error {
	taskPath := filepath.Join(r.RepoRoot, TaskRelPath)

	for {
		taskFile, err := r.loadOrPlanTask(ctx, backend, models, taskPath)
		if err != nil {
			return err
		}
		r.TaskFile = taskFile

		if taskFile.Task.Status == tasks.StatusDone {
			entry := fmt.Sprintf("- %s %s %s - done (no remaining work)", timeNowUTC(), taskFile.Task.ID, taskFile.Task.Title)
			if err := AppendProgress(r.RepoRoot, entry); err != nil {
				return err
			}
			if _, err := ArchiveTaskFile(r.RepoRoot, taskFile); err != nil {
				return err
			}
			_ = os.Remove(taskPath)
			break
		}

		r.Resume = false
		err = r.ExecuteTaskWithTask(ctx, backend, &taskFile.Task, models.Fast.Name, models.Fast.Variant)
		if err != nil {
			return err
		}
	}

	if err := state.Clear(r.RepoRoot); err != nil {
		return fmt.Errorf("clear state: %w", err)
	}

	return nil
}

func (r *Runner) loadOrPlanTask(ctx context.Context, backend relay.Provider, models Models, taskPath string) (*tasks.TaskFile, error) {
	if r.Resume && r.State.ActiveTaskID != "" {
		return tasks.LoadTaskFile(taskPath)
	}

	if _, err := os.Stat(taskPath); err == nil {
		return tasks.LoadTaskFile(taskPath)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat task file: %w", err)
	}

	planner := decomposer.New(backend, r.RepoRoot)
	if err := planner.PlanNext(ctx, r.PRDPath, r.ProgressPath, PlanOptionsFromModels(models)); err != nil {
		return nil, err
	}

	return tasks.LoadTaskFile(taskPath)
}

func PlanOptionsFromModels(models Models) decomposer.PlanOptions {
	return decomposer.PlanOptions{
		FastModel:   models.Fast.Name,
		FastVariant: models.Fast.Variant,
		SlowModel:   models.Slow.Name,
		SlowVariant: models.Slow.Variant,
	}
}

func timeNowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
