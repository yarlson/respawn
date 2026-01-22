package run

import (
	"context"
	"fmt"
	"github.com/yarlson/respawn/internal/gitx"
	"github.com/yarlson/respawn/internal/state"
	"github.com/yarlson/respawn/internal/tasks"
)

// RetryPolicy manages the lifecycle of a task execution with retries and resets.
type RetryPolicy struct {
	MaxAttempts int
	MaxCycles   int
}

func (p *RetryPolicy) Execute(ctx context.Context, r *Runner, task *tasks.Task, execute func(ctx context.Context, sessionID string) error) error {
	// Initialize or resume state
	if r.State.ActiveTaskID != task.ID {
		r.State.ActiveTaskID = task.ID
		r.State.Cycle = 1
		r.State.Attempt = 1
		// Store last save point if not already set
		if r.State.LastSavepointCommit == "" {
			hash, err := gitx.CurrentHash(ctx, r.RepoRoot)
			if err != nil {
				return fmt.Errorf("get current hash for savepoint: %w", err)
			}
			r.State.LastSavepointCommit = hash
		}
		if err := state.Save(r.RepoRoot, r.State); err != nil {
			return err
		}
	}

	for r.State.Cycle <= p.MaxCycles {
		for r.State.Attempt <= p.MaxAttempts {
			fmt.Printf("Attempt: cycle %d, attempt %d/%d\n", r.State.Cycle, r.State.Attempt, p.MaxAttempts)

			err := execute(ctx, r.State.BackendSessionID)
			if err == nil {
				// Success!
				return nil
			}

			fmt.Printf("Attempt failed: %v\n", err)

			if r.State.Attempt < p.MaxAttempts {
				r.State.Attempt++
				if err := state.Save(r.RepoRoot, r.State); err != nil {
					return err
				}
			} else {
				break
			}
		}

		// Cycle exhausted, increment cycle and reset
		if r.State.Cycle < p.MaxCycles {
			r.State.Cycle++
			fmt.Printf("Cycle %d exhausted. Resetting to last save point: %s\n", r.State.Cycle-1, r.State.LastSavepointCommit)
			if err := gitx.ResetHard(ctx, r.RepoRoot, r.State.LastSavepointCommit); err != nil {
				return fmt.Errorf("reset to save point: %w", err)
			}
			r.State.Attempt = 1
			r.State.BackendSessionID = "" // Force new session
			if err := state.Save(r.RepoRoot, r.State); err != nil {
				return err
			}
		} else {
			break
		}
	}

	// All cycles exhausted
	task.Status = tasks.StatusFailed
	return fmt.Errorf("task %s failed after %d cycles", task.ID, p.MaxCycles)
}
