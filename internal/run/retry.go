package run

import (
	"context"
	"fmt"

	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/state"
	"github.com/yarlson/turbine/internal/tasks"
	"github.com/yarlson/turbine/internal/ui"
)

// RetryPolicy manages the lifecycle of a task execution with retries and resets.
type RetryPolicy struct {
	MaxStrokes   int
	MaxRotations int
}

func (p *RetryPolicy) Execute(ctx context.Context, r *Runner, task *tasks.Task, execute func(ctx context.Context, sessionID string) error) error {
	// Initialize or resume state
	if r.State.ActiveTaskID != task.ID {
		r.State.ActiveTaskID = task.ID
		r.State.Rotation = 1
		r.State.Stroke = 1
		// Store last save point if not already set
		if r.State.LastSavepointCommit == "" {
			hash, err := gitx.CurrentHash(ctx, r.RepoRoot)
			if err != nil {
				return fmt.Errorf("get commit hash: %w", err)
			}
			r.State.LastSavepointCommit = hash
		}
		if err := state.Save(r.RepoRoot, r.State); err != nil {
			return err
		}
	}

	for r.State.Rotation <= p.MaxRotations {
		for r.State.Stroke <= p.MaxStrokes {
			fmt.Printf("  %s Stroke %d/%d (rotation %d)\n", ui.InProgressMarker(), r.State.Stroke, p.MaxStrokes, r.State.Rotation)

			err := execute(ctx, r.State.BackendSessionID)
			if err == nil {
				// Success!
				return nil
			}

			fmt.Printf("  %s %v\n", ui.FailureMarker(), ui.Dim(fmt.Sprintf("Stroke failed: %v", err)))

			if r.State.Stroke < p.MaxStrokes {
				r.State.Stroke++
				if err := state.Save(r.RepoRoot, r.State); err != nil {
					return err
				}
			} else {
				break
			}
		}

		// Rotation exhausted, increment rotation and reset
		if r.State.Rotation < p.MaxRotations {
			r.State.Rotation++
			fmt.Printf("  %s Rotation %d failed. Re-spinning at %s\n", ui.Yellow("⟳"), r.State.Rotation-1, ui.Dim(r.State.LastSavepointCommit[:8]))
			if err := gitx.ResetHard(ctx, r.RepoRoot, r.State.LastSavepointCommit); err != nil {
				return fmt.Errorf("reset to savepoint: %w", err)
			}
			r.State.Stroke = 1
			r.State.BackendSessionID = "" // Force new session
			if err := state.Save(r.RepoRoot, r.State); err != nil {
				return err
			}
		} else {
			break
		}
	}

	// All rotations exhausted
	task.Status = tasks.StatusFailed
	return fmt.Errorf("%s failed after %d rotations", task.ID, p.MaxRotations)
}
