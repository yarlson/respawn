package decomposer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/prompt"
	"github.com/yarlson/turbine/internal/prompt/roles"
	"github.com/yarlson/turbine/internal/tasks"
)

type Decomposer struct {
	backend  relay.Provider
	repoRoot string
}

// DecomposeOptions configures the two-phase decomposition process.
type DecomposeOptions struct {
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

// Decompose instructs the coding agent to create .turbine/tasks.yaml from a PRD.
// It uses a two-phase approach:
// Phase 1 (fast model): Explore the codebase to understand patterns and conventions
// Phase 2 (slow model): Generate the tasks.yaml using the gathered context
// Both phases run in the same session using --continue.
func (d *Decomposer) Decompose(ctx context.Context, prdPath string, opts DecomposeOptions) error {
	prdContent, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	outputPath := ".turbine/tasks.yaml"
	tasksPath := filepath.Join(d.repoRoot, outputPath)

	// Phase 1: Explore - no methodologies needed
	exploreCtx := prompt.ExecutionContext{Phase: prompt.PhaseExplore}
	exploreMethods := prompt.SelectMethodologies(exploreCtx)
	exploreSystemPrompt := prompt.Compose(roles.RoleExplorer, exploreMethods, "")
	explorePrompt := exploreSystemPrompt + "\n\n" + prompt.ExploreUserPrompt(string(prdContent))

	// Phase 2: Decompose - uses Planning methodology
	decomposeCtx := prompt.ExecutionContext{Phase: prompt.PhaseDecompose}
	decomposeMethods := prompt.SelectMethodologies(decomposeCtx)
	decomposeSystemPrompt := prompt.Compose(roles.RoleDecomposer, decomposeMethods, "")
	decomposePrompt := decomposeSystemPrompt + "\n\n" + prompt.DecomposeUserPrompt(string(prdContent), outputPath)

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
						Prompt:  decomposePrompt,
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
		lastErr = d.validateTasksFile(tasksPath)
		if lastErr == nil {
			return nil
		}

		if i >= maxValidationRetries {
			break
		}

		fileContent, _ := os.ReadFile(tasksPath)
		fixPrompt := prompt.DecomposeFixPrompt(string(prdContent), string(fileContent), lastErr.Error())

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

	return fmt.Errorf("decompose failed after %d retries: %w", maxValidationRetries, lastErr)
}

func (d *Decomposer) runWorkflow(ctx context.Context, exec *relay.Executor, workflow *relay.Workflow) error {
	events := make(chan relay.Event, 128)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range events {
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

// validateTasksFile checks that the tasks file exists and is valid.
func (d *Decomposer) validateTasksFile(tasksPath string) error {
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tasks file was not created at %s", tasksPath)
		}
		return fmt.Errorf("read tasks file: %w", err)
	}

	if len(content) == 0 {
		return fmt.Errorf("tasks file is empty")
	}

	// Parse and validate
	taskList, err := tasks.Load(tasksPath)
	if err != nil {
		return fmt.Errorf("parse tasks: %w", err)
	}

	if err := taskList.Validate(); err != nil {
		return fmt.Errorf("validate tasks: %w", err)
	}

	return nil
}
