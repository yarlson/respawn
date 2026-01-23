package decomposer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/prompt"
	"github.com/yarlson/turbine/internal/tasks"
)

type Decomposer struct {
	backend  backends.Backend
	repoRoot string
}

// DecomposeOptions configures the two-phase decomposition process.
type DecomposeOptions struct {
	// FastModel is used for Phase 1: codebase exploration
	FastModel string
	// SlowModel is used for Phase 2: task generation
	SlowModel string
	// ArtifactsDir is the path where stdout/stderr and other run data should be captured
	ArtifactsDir string
}

const maxValidationRetries = 2

func New(backend backends.Backend, repoRoot string) *Decomposer {
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

	// Start session without specifying model - we'll set it per-send
	sessionID, err := d.backend.StartSession(ctx, backends.SessionOptions{
		WorkingDir:   d.repoRoot,
		ArtifactsDir: opts.ArtifactsDir,
	})
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	// Phase 1: Explore codebase with fast model
	explorePrompt := prompt.ExploreSystemPrompt + "\n\n" + prompt.ExploreUserPrompt(string(prdContent))
	_, err = d.backend.Send(ctx, sessionID, explorePrompt, backends.SendOptions{
		Model: opts.FastModel,
	})
	if err != nil {
		return fmt.Errorf("explore phase: %w", err)
	}

	// Phase 2: Generate tasks with slow model (continues in same session)
	decomposePrompt := prompt.DecomposerSystemPrompt + "\n\n" + prompt.DecomposeUserPrompt(string(prdContent), outputPath)

	var lastErr error

	for i := 0; i <= maxValidationRetries; i++ {
		_, err := d.backend.Send(ctx, sessionID, decomposePrompt, backends.SendOptions{
			Model: opts.SlowModel,
		})
		if err != nil {
			return fmt.Errorf("send prompt: %w", err)
		}

		// Validate the file the agent wrote
		lastErr = d.validateTasksFile(tasksPath)
		if lastErr == nil {
			return nil
		}

		// Ask agent to fix the file
		if i < maxValidationRetries {
			fileContent, _ := os.ReadFile(tasksPath)
			decomposePrompt = prompt.DecomposeFixPrompt(string(prdContent), string(fileContent), lastErr.Error())
		}
	}

	return fmt.Errorf("decompose failed after %d retries: %w", maxValidationRetries, lastErr)
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
