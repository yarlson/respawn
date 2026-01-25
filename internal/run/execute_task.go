package run

import (
	"context"
	"fmt"
	"path/filepath"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/gitx"
	"github.com/yarlson/turbine/internal/prompt"
	"github.com/yarlson/turbine/internal/prompt/roles"
	relaystore "github.com/yarlson/turbine/internal/relay/store"
	"github.com/yarlson/turbine/internal/tasks"
	"github.com/yarlson/turbine/internal/ui"
)

// ExecuteTask selects and executes the next runnable task using the provided model and variant.
func (r *Runner) ExecuteTask(ctx context.Context, backend relay.Provider, model, variant string) error {
	task := r.NextRunnableTask()
	if task == nil {
		return fmt.Errorf("no runnable tasks found")
	}
	return r.ExecuteTaskWithTask(ctx, backend, task, model, variant)
}

// ExecuteTaskWithTask executes a single task with the given model and variant.
func (r *Runner) ExecuteTaskWithTask(ctx context.Context, backend relay.Provider, task *tasks.Task, model, variant string) error {
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
		phase := prompt.PhaseImplement
		if r.State.Stroke > 1 || r.State.Rotation > 1 {
			phase = prompt.PhaseRetry
		}

		// Build execution context
		execCtx := prompt.ExecutionContext{
			Phase:    phase,
			Attempt:  r.State.Stroke,
			Rotation: r.State.Rotation,
		}

		// Select role based on phase
		role := roles.RoleImplementer
		if phase == prompt.PhaseRetry {
			role = roles.RoleRetrier
		}

		// Compose system prompt with methodologies
		meths := prompt.SelectMethodologies(execCtx)
		systemPrompt := prompt.Compose(role, meths, "")

		// Build user prompt based on phase
		var userPrompt string
		if phase == prompt.PhaseRetry {
			userPrompt = prompt.RetryUserPrompt(*task, lastFailureOutput)
		} else {
			userPrompt = prompt.ImplementUserPrompt(*task)
		}

		// Combine system and user prompts
		fullPrompt := systemPrompt + "\n\n---\n\n" + userPrompt

		workflowID := fmt.Sprintf("%s-%s", r.State.RunID, task.ID)
		exec := relay.NewExecutor(backend, relay.WithStore(relaystore.New(r.RepoRoot)))
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

		if err := runWorkflow(ctx, exec, workflow); err != nil {
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

func runWorkflow(ctx context.Context, exec *relay.Executor, workflow *relay.Workflow) error {
	events := make(chan relay.Event, 256)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range events {
		}
	}()

	_, err := exec.Run(ctx, workflow, events)
	close(events)
	<-done
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
