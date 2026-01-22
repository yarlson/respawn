package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"respawn/internal/config"
	"respawn/internal/gitx"
	"respawn/internal/state"
	"respawn/internal/tasks"
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

	// tasks.yaml presence - check this early as it defines if we are in a respawn context
	tasksPath := filepath.Join(repoRoot, ".respawn", "tasks.yaml")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".respawn/tasks.yaml missing at %s", tasksPath)
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

	// Clean working tree rule
	// IMPORTANT: Check this BEFORE we modify .gitignore ourselves
	if !exists {
		dirty, err := gitx.IsDirty(ctx, repoRoot)
		if err != nil {
			return nil, fmt.Errorf("check if repo is dirty: %w", err)
		}
		if dirty {
			return nil, fmt.Errorf("working tree is dirty; commit or stash changes before starting a new run")
		}
	}

	// .gitignore check
	missing, err := gitx.MissingRespawnIgnores(ctx, repoRoot)
	if err != nil {
		return nil, fmt.Errorf("check gitignore: %w", err)
	}

	if len(missing) > 0 {
		if cfg.AutoAddIgnore {
			if err := gitx.AddIgnoresToGitignore(repoRoot, missing); err != nil {
				return nil, fmt.Errorf("add missing ignores to .gitignore: %w", err)
			}
		} else {
			fmt.Printf("Warning: The following paths are not ignored by git:\n")
			for _, m := range missing {
				fmt.Printf("  - %s\n", m)
			}
			fmt.Printf("Run with --yes to automatically add them to .gitignore.\n")
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

	fmt.Printf("Tasks: %d total, %d done, %d runnable, %d blocked, %d failed\n",
		total, done, runnable, blocked, failed)
}
