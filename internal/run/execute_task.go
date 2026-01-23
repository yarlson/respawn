package run

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/prompt"
	"github.com/yarlson/turbine/internal/tasks"
	"github.com/yarlson/turbine/internal/ui"
)

// ExecuteTask selects and executes the next runnable task.
func (r *Runner) ExecuteTask(ctx context.Context, backend backends.Backend) error {
	task := r.NextRunnableTask()
	if task == nil {
		return fmt.Errorf("no runnable tasks found")
	}
	return r.ExecuteTaskWithTask(ctx, backend, task)
}

// ExecuteTaskWithTask executes a specific task.
func (r *Runner) ExecuteTaskWithTask(ctx context.Context, backend backends.Backend, task *tasks.Task) error {
	fmt.Printf("%s %s\n", ui.Section("›", ui.Bold(task.Title)), ui.Dim(fmt.Sprintf("[%s]", task.ID)))

	policy := &RetryPolicy{
		MaxStrokes:   r.Config.Retry.Strokes,
		MaxRotations: r.Config.Retry.Rotations,
	}

	// Artifacts setup
	arts, err := NewArtifacts(r.RepoRoot, r.State.RunID)
	if err != nil {
		return fmt.Errorf("set up artifacts: %w", err)
	}

	err = policy.Execute(ctx, r, task, func(ctx context.Context, sessionID string) error {
		// Session setup
		if sessionID == "" {
			var err error
			sessionID, err = backend.StartSession(ctx, backends.SessionOptions{
				WorkingDir:   r.RepoRoot,
				ArtifactsDir: arts.Root(),
			})
			if err != nil {
				return fmt.Errorf("start session: %w", err)
			}
			r.State.BackendSessionID = sessionID
			fmt.Printf("  %s %s\n", ui.Dim("Session:"), ui.Dim(sessionID))
		}

		// Prompt building
		userPrompt := prompt.ImplementUserPrompt(*task)

		// Invoke backend
		_, err = backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
		if err != nil {
			return fmt.Errorf("backend failed: %w", err)
		}

		// Run verify commands
		fmt.Printf("  %s\n", ui.InProgressMarker()+" Verifying...")
		_, verifyErr := RunVerification(ctx, arts, task.Verify)
		if verifyErr != nil {
			fmt.Printf("  %s\n", ui.FailureMarker()+" Verification failed")
			return fmt.Errorf("verification failed: %w", verifyErr)
		}
		fmt.Printf("  %s\n", ui.SuccessMarker()+" Verification passed")
		return nil
	})

	// Save task status (either Done if err == nil, or Failed if policy returned error)
	if err == nil {
		task.Status = tasks.StatusDone
	} else {
		// policy.Execute already sets task.Status = tasks.StatusFailed on exhaustion
		fmt.Printf("%s %s\n  %s\n", ui.FailureMarker(), ui.Bold(task.Title), ui.Red(fmt.Sprintf("Failed after max rotations: %v", err)))
	}

	tasksPath := filepath.Join(r.RepoRoot, ".turbine", "tasks.yaml")
	if saveErr := r.Tasks.Save(tasksPath); saveErr != nil {
		return fmt.Errorf("save tasks: %w", saveErr)
	}

	if err == nil {
		// Commit changes on success
		footer := fmt.Sprintf("Turbine: %s", task.ID)
		hash, commitErr := gitx.CommitSavePoint(ctx, r.RepoRoot, task.CommitMessage, footer)
		if commitErr != nil {
			return fmt.Errorf("commit changes: %w", commitErr)
		}
		fmt.Printf("  %s %s\n", ui.SuccessMarker(), ui.Dim(hash))

		// Update last save point for next task
		r.State.LastSavepointCommit = hash
		r.State.ActiveTaskID = "" // Reset for next task
		r.State.BackendSessionID = ""
	}

	return err
}

// NextRunnableTask returns the first task that is 'todo' and has all dependencies 'done'.
func (r *Runner) NextRunnableTask() *tasks.Task {
	// Map tasks by ID for dependency checking
	taskMap := make(map[string]tasks.Task)
	for _, t := range r.Tasks.Tasks {
		taskMap[t.ID] = t
	}

	for i, t := range r.Tasks.Tasks {
		if t.Status != tasks.StatusTodo {
			continue
		}

		runnable := true
		for _, depID := range t.Deps {
			dep, ok := taskMap[depID]
			if !ok || dep.Status != tasks.StatusDone {
				runnable = false
				break
			}
		}

		if runnable {
			return &r.Tasks.Tasks[i]
		}
	}
	return nil
}

// FindTaskByID returns a task by its ID.
func (r *Runner) FindTaskByID(id string) *tasks.Task {
	for i, t := range r.Tasks.Tasks {
		if t.ID == id {
			return &r.Tasks.Tasks[i]
		}
	}
	return nil
}
