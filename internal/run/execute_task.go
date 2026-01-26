package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/gitx"
	filestore "github.com/yarlson/turbine/internal/relay/store"
	"github.com/yarlson/turbine/internal/relay/stream"
	"github.com/yarlson/turbine/internal/tasks"
	"github.com/yarlson/turbine/internal/ui"
)

// ExecuteTask executes the currently loaded task.
func (r *Runner) ExecuteTask(ctx context.Context, backend relay.Provider, model, variant string) error {
	if r.TaskFile == nil {
		return fmt.Errorf("no active task loaded")
	}
	return r.ExecuteTaskWithTask(ctx, backend, &r.TaskFile.Task, model, variant)
}

// ExecuteTaskWithTask executes a single task with the given model and variant.
func (r *Runner) ExecuteTaskWithTask(ctx context.Context, backend relay.Provider, task *tasks.Task, model, variant string) error {
	if r.TaskFile == nil {
		return fmt.Errorf("no active task file loaded")
	}

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

	// Track failure output for retry context
	var lastFailureOutput string

	err = policy.Execute(ctx, r, task, func(ctx context.Context) error {
		// Determine phase based on current stroke and rotation
		isRetry := r.State.Stroke > 1 || r.State.Rotation > 1
		execCtx := promptContext{
			IsRetry:  isRetry,
			Attempt:  r.State.Stroke,
			Rotation: r.State.Rotation,
		}

		// Build user prompt based on phase
		var userPrompt string
		if isRetry {
			userPrompt = retryUserPrompt(*task, lastFailureOutput)
		} else {
			userPrompt = implementUserPrompt(*task)
		}

		// Combine system and user prompts
		fullPrompt := buildTaskPrompt(execCtx, userPrompt)

		workflowID := fmt.Sprintf("%s-%s", r.State.RunID, task.ID)
		store := filestore.New(r.RepoRoot)
		exec := relay.NewExecutor(backend, relay.WithStore(store))
		workflow := &relay.Workflow{
			ID:         workflowID,
			WorkingDir: r.RepoRoot,
			Model:      model,
			Variant:    variant,
			Sessions: []relay.Session{
				{
					Steps: []relay.Step{
						{
							Prompt:   fullPrompt,
							Continue: r.State.Stroke > 1,
							PostHook: func(_ *relay.StepContext, _ *relay.StepResult) error {
								fmt.Printf("  %s\n", ui.InProgressMarker()+" Verifying...")
								_, verifyErr := RunVerification(ctx, arts, task.Verify)
								if verifyErr != nil {
									fmt.Printf("  %s\n", ui.FailureMarker()+" Verification failed")
									lastFailureOutput = verifyErr.Error()
									return fmt.Errorf("verification failed: %w", verifyErr)
								}
								fmt.Printf("  %s\n", ui.SuccessMarker()+" Verification passed")
								return nil
							},
						},
					},
				},
			},
		}

		if err := runWorkflow(ctx, exec, workflow, store); err != nil {
			return fmt.Errorf("backend failed: %w", err)
		}

		return nil
	})

	// Save task status (either Done if err == nil, or Failed if policy returned error)
	if err == nil {
		task.Status = tasks.StatusDone
	} else {
		// policy.Execute already sets task.Status = tasks.StatusFailed on exhaustion
		fmt.Printf("%s %s\n  %s\n", ui.FailureMarker(), ui.Bold(task.Title), ui.Red(fmt.Sprintf("Failed after max rotations: %v", err)))
	}

	taskPath := filepath.Join(r.RepoRoot, TaskRelPath)
	if saveErr := r.TaskFile.Save(taskPath); saveErr != nil {
		return fmt.Errorf("save task: %w", saveErr)
	}

	if err == nil {
		// Commit changes on success
		footer := fmt.Sprintf("Turbine: %s", task.ID)
		hash, commitErr := gitx.CommitSavePoint(ctx, r.RepoRoot, task.CommitMessage, footer)
		if commitErr != nil {
			return fmt.Errorf("commit changes: %w", commitErr)
		}
		fmt.Printf("  %s %s\n", ui.SuccessMarker(), ui.Dim(hash))

		entry := fmt.Sprintf("- %s %s %s - done (commit %s)", timeNowUTC(), task.ID, task.Title, hash)
		if err := AppendProgress(r.RepoRoot, entry); err != nil {
			return err
		}
		if _, err := ArchiveTaskFile(r.RepoRoot, r.TaskFile); err != nil {
			return err
		}
		if err := os.Remove(taskPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove task file: %w", err)
		}

		// Update last save point for next task
		r.State.LastSavepointCommit = hash
		r.State.ActiveTaskID = "" // Reset for next task
		r.State.BackendSessionID = ""
	}

	if err != nil {
		entry := fmt.Sprintf("- %s %s %s - failed", timeNowUTC(), task.ID, task.Title)
		if progressErr := AppendProgress(r.RepoRoot, entry); progressErr != nil {
			return progressErr
		}
	}

	return err
}

func runWorkflow(ctx context.Context, exec *relay.Executor, workflow *relay.Workflow, store *filestore.FileStore) error {
	events := make(chan relay.Event, 256)
	done := make(chan struct{})
	go func() {
		defer close(done)
		if store == nil || workflow.ID == "" {
			for range events {
			}
			return
		}
		for evt := range events {
			stream.AppendEvent(ctx, store, workflow.ID, evt)
		}
	}()

	_, err := exec.Run(ctx, workflow, events)
	close(events)
	<-done
	return err
}
