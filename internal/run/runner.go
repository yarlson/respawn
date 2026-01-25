package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/state"
	"github.com/yarlson/turbine/internal/tasks"
	"github.com/yarlson/turbine/internal/ui"
)

type Runner struct {
	RepoRoot string
	Tasks    *tasks.TaskList
	State    *state.RunState
	Config   config.Defaults
	Resume   bool
}

type Config struct {
	AutoAddIgnore bool
	Cwd           string
	Defaults      config.Defaults
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

	// tasks.yaml presence - check this early as it defines if we are in a turbine context
	tasksPath := filepath.Join(repoRoot, ".turbine", "tasks.yaml")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("task manifest not found: %s", tasksPath)
	}

	taskList, err := tasks.Load(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

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

	return &Runner{
		RepoRoot: repoRoot,
		Tasks:    taskList,
		State:    runState,
		Config:   cfg.Defaults,
		Resume:   exists,
	}, nil
}

func (r *Runner) PrintSummary() {
	var done, todo, failed, runnable, blocked int
	total := len(r.Tasks.Tasks)

	// Map tasks by ID for easy dependency checking
	taskMap := make(map[string]tasks.Task)
	for _, t := range r.Tasks.Tasks {
		taskMap[t.ID] = t
	}

	for _, t := range r.Tasks.Tasks {
		switch t.Status {
		case tasks.StatusDone:
			done++
		case tasks.StatusFailed:
			failed++
		case tasks.StatusTodo:
			todo++
			// Check if runnable or blocked
			isBlocked := false
			for _, depID := range t.Deps {
				dep, ok := taskMap[depID]
				if !ok || dep.Status != tasks.StatusDone {
					isBlocked = true
					break
				}
			}
			if isBlocked {
				blocked++
			} else {
				runnable++
			}
		}
	}

	fmt.Printf("\n%s\n", ui.Divider(40))
	fmt.Printf("%d tasks: %s %s %s %s\n",
		total,
		ui.Green(fmt.Sprintf("%d cleared", done)),
		ui.Cyan(fmt.Sprintf("%d ready", runnable)),
		ui.Yellow(fmt.Sprintf("%d blocked", blocked)),
		ui.Red(fmt.Sprintf("%d failed", failed)),
	)
}

// Run executes tasks from the manifest using the provided backend, model, and variant.
func (r *Runner) Run(ctx context.Context, backend relay.Provider, model, variant string) error {
	for {
		// 1. Check if we have an active task to resume
		var task *tasks.Task
		if r.Resume && r.State.ActiveTaskID != "" {
			task = r.FindTaskByID(r.State.ActiveTaskID)
			// If for some reason the task is not found or already done, fall back to next runnable
			if task == nil || task.Status != tasks.StatusTodo {
				task = r.NextRunnableTask()
			}
		} else {
			task = r.NextRunnableTask()
		}

		if task == nil {
			break
		}

		// Ensure we don't try to resume multiple times if we're in the loop
		r.Resume = false

		// ExecuteTask handles its own retries/rotations and saves tasks.yaml
		_ = r.ExecuteTaskWithTask(ctx, backend, task, model, variant)
	}

	r.PrintSummary()

	// Exit code handling logic
	var failed int
	for _, t := range r.Tasks.Tasks {
		if t.Status == tasks.StatusFailed {
			failed++
		}
	}

	// Normal completion - clear state
	if err := state.Clear(r.RepoRoot); err != nil {
		return fmt.Errorf("clear state: %w", err)
	}

	if failed > 0 {
		return fmt.Errorf("%d tasks failed", failed)
	}

	return nil
}
