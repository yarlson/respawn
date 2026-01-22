package run

import (
	"context"
	"fmt"
	"path/filepath"
	"respawn/internal/backends"
	"respawn/internal/gitx"
	"respawn/internal/prompt"
	"respawn/internal/tasks"
)

// ExecuteTask selects and executes the next runnable task.
func (r *Runner) ExecuteTask(ctx context.Context, backend backends.Backend) error {
	task := r.NextRunnableTask()
	if task == nil {
		return fmt.Errorf("no runnable tasks found")
	}

	fmt.Printf("TASK: %s (%s) start\n", task.Title, task.ID)

	// Artifacts setup
	arts, err := NewArtifacts(r.RepoRoot, r.State.RunID)
	if err != nil {
		return fmt.Errorf("create artifacts: %w", err)
	}

	// Session setup
	sessionID, err := backend.StartSession(ctx, backends.SessionOptions{
		WorkingDir:   r.RepoRoot,
		ArtifactsDir: arts.Root(),
	})
	if err != nil {
		return fmt.Errorf("start backend session: %w", err)
	}
	fmt.Printf("Session ID: %s\n", sessionID)

	// Prompt building
	userPrompt := prompt.ImplementUserPrompt(*task)

	// Invoke backend
	_, err = backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
	if err != nil {
		return fmt.Errorf("backend execution failed: %w", err)
	}

	// Run verify commands
	fmt.Printf("Verification: running...\n")
	_, verifyErr := RunVerification(ctx, arts, task.Verify)
	if verifyErr != nil {
		fmt.Printf("Verification: FAILED\n")
		return fmt.Errorf("verification failed for task %s: %w. See logs in %s", task.ID, verifyErr, arts.Root())
	}
	fmt.Printf("Verification: PASSED\n")

	// Update task status
	task.Status = tasks.StatusDone
	tasksPath := filepath.Join(r.RepoRoot, ".respawn", "tasks.yaml")
	if err := r.Tasks.Save(tasksPath); err != nil {
		return fmt.Errorf("save tasks: %w", err)
	}

	// Commit changes
	footer := fmt.Sprintf("Respawn: %s", task.ID)
	hash, err := gitx.CommitSavePoint(ctx, r.RepoRoot, task.CommitMessage, footer)
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	fmt.Printf("Commit: %s\n", hash)

	return nil
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
