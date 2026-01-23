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

type Backend interface {
	backends.Backend
}

type Decomposer struct {
	Backend  Backend
	RepoRoot string
}

const maxValidationRetries = 2

func New(backend backends.Backend, repoRoot string) *Decomposer {
	return &Decomposer{
		Backend:  backend,
		RepoRoot: repoRoot,
	}
}

// Decompose instructs the coding agent to create .turbine/tasks.yaml from a PRD.
// The agent writes the file directly using its tools.
func (d *Decomposer) Decompose(ctx context.Context, prdPath string, artifactsDir string) error {
	prdContent, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	outputPath := ".turbine/tasks.yaml"
	tasksPath := filepath.Join(d.RepoRoot, outputPath)

	sessionID, err := d.Backend.StartSession(ctx, backends.SessionOptions{
		WorkingDir:   d.RepoRoot,
		ArtifactsDir: artifactsDir,
	})
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	// Instruct the coding agent to create the tasks file
	userPrompt := prompt.DecomposerSystemPrompt + "\n\n" + prompt.DecomposeUserPrompt(string(prdContent), outputPath)

	var lastErr error

	for i := 0; i <= maxValidationRetries; i++ {
		_, err := d.Backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
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
			userPrompt = prompt.DecomposeFixPrompt(string(prdContent), string(fileContent), lastErr.Error())
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
