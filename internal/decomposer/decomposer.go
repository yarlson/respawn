package decomposer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	relay "github.com/yarlson/relay"
	filestore "github.com/yarlson/turbine/internal/relay/store"
	"github.com/yarlson/turbine/internal/relay/stream"
	"github.com/yarlson/turbine/internal/tasks"
)

type Decomposer struct {
	backend  relay.Provider
	repoRoot string
}

// PlanOptions configures the two-phase planning process.
type PlanOptions struct {
	// FastModel is used for Phase 1: codebase exploration
	FastModel   string
	FastVariant string
	// SlowModel is used for Phase 2: task generation
	SlowModel   string
	SlowVariant string
	// ArtifactsDir is the path where stdout/stderr and other run data should be captured
	ArtifactsDir string
}

const maxValidationRetries = 2

func New(backend relay.Provider, repoRoot string) *Decomposer {
	return &Decomposer{
		backend:  backend,
		repoRoot: repoRoot,
	}
}

// PlanNext instructs the coding agent to create .turbine/task.yaml from a PRD and progress.
// It uses a two-phase approach:
// Phase 1 (fast model): Explore the codebase and progress to understand context
// Phase 2 (slow model): Generate the next task using the gathered context
// Both phases run in the same session using --continue.
func (d *Decomposer) PlanNext(ctx context.Context, prdPath, progressPath string, opts PlanOptions) error {
	prdContent, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	progressContent := ""
	if progressPath != "" {
		content, err := os.ReadFile(progressPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read progress: %w", err)
		}
		if err == nil {
			progressContent = string(content)
		}
	}

	outputPath := ".turbine/task.yaml"
	taskPath := filepath.Join(d.repoRoot, outputPath)

	// Phase 1: Explore - no methodologies needed
	explorePrompt := buildExplorePrompt(string(prdContent), progressContent)

	// Phase 2: Plan - uses Planning methodology
	planPrompt := buildPlanPrompt(string(prdContent), progressContent, outputPath)

	exec := relay.NewExecutor(d.backend)

	if err := d.runWorkflow(ctx, exec, &relay.Workflow{
		WorkingDir: d.repoRoot,
		Sessions: []relay.Session{
			{
				Steps: []relay.Step{
					{
						Prompt:  explorePrompt,
						Model:   opts.FastModel,
						Variant: opts.FastVariant,
					},
					{
						Prompt:  planPrompt,
						Model:   opts.SlowModel,
						Variant: opts.SlowVariant,
					},
				},
			},
		},
	}); err != nil {
		return err
	}

	var lastErr error
	for i := 0; i <= maxValidationRetries; i++ {
		// Validate the file the agent wrote
		lastErr = d.validateTaskFile(taskPath)
		if lastErr == nil {
			return nil
		}

		if i >= maxValidationRetries {
			break
		}

		fileContent, _ := os.ReadFile(taskPath)
		fixPrompt := buildPlanFixPrompt(string(prdContent), progressContent, string(fileContent), lastErr.Error())

		if err := d.runWorkflow(ctx, exec, &relay.Workflow{
			WorkingDir: d.repoRoot,
			Sessions: []relay.Session{
				{
					Steps: []relay.Step{
						{
							Prompt:   fixPrompt,
							Model:    opts.SlowModel,
							Variant:  opts.SlowVariant,
							Continue: true,
						},
					},
				},
			},
		}); err != nil {
			return err
		}
	}

	return fmt.Errorf("plan failed after %d retries: %w", maxValidationRetries, lastErr)
}

func (d *Decomposer) runWorkflow(ctx context.Context, exec *relay.Executor, workflow *relay.Workflow) error {
	events := make(chan relay.Event, 128)
	done := make(chan struct{})
	go func() {
		defer close(done)
		if workflow.ID == "" {
			for range events {
			}
			return
		}
		store := filestore.New(d.repoRoot)
		for evt := range events {
			stream.AppendEvent(ctx, store, workflow.ID, evt)
		}
	}()

	_, err := exec.Run(ctx, workflow, events)
	close(events)
	<-done
	if err != nil {
		return fmt.Errorf("run workflow: %w", err)
	}
	return nil
}

// validateTaskFile checks that the task file exists and is valid.
func (d *Decomposer) validateTaskFile(taskPath string) error {
	content, err := os.ReadFile(taskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("task file was not created at %s", taskPath)
		}
		return fmt.Errorf("read task file: %w", err)
	}

	if len(content) == 0 {
		return fmt.Errorf("task file is empty")
	}

	// Parse and validate
	taskFile, err := tasks.LoadTaskFile(taskPath)
	if err != nil {
		return fmt.Errorf("parse task: %w", err)
	}

	if err := taskFile.Validate(); err != nil {
		return fmt.Errorf("validate task: %w", err)
	}

	return nil
}
