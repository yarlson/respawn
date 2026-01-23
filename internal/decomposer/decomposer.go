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

	// Combine system prompt with user prompt for backends that don't support separate system prompts
	userPrompt := prompt.DecomposerSystemPrompt + "\n\n" + prompt.DecomposeUserPrompt(string(prdContent), outputPath)

	var lastErr error
	var taskList *tasks.TaskList

	for i := 0; i <= maxValidationRetries; i++ {
		res, err := d.Backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
		if err != nil {
			return fmt.Errorf("send prompt: %w", err)
		}

		// First, try to extract YAML from the backend's text output
		tasksYAML := extractYAML(res.Output)

		// If no YAML in output, check if the backend wrote the file directly (e.g., OpenCode uses tools)
		if tasksYAML == "" {
			if fileContent, readErr := os.ReadFile(tasksPath); readErr == nil && len(fileContent) > 0 {
				tasksYAML = string(fileContent)
			}
		}

		if tasksYAML == "" {
			lastErr = fmt.Errorf("no YAML found in backend response or output file")
		} else {
			taskList, lastErr = validateTasksYAML(tasksYAML)
			if lastErr == nil {
				break
			}
		}

		if i < maxValidationRetries {
			userPrompt = prompt.DecomposeFixPrompt(string(prdContent), tasksYAML, lastErr.Error())
		}
	}

	if lastErr != nil {
		return fmt.Errorf("decompose failed after %d retries: %w", maxValidationRetries, lastErr)
	}

	// Ensure directory exists and save the validated taskList
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0755); err != nil {
		return fmt.Errorf("create .turbine dir: %w", err)
	}

	if err := taskList.Save(tasksPath); err != nil {
		return fmt.Errorf("save tasks: %w", err)
	}

	return nil
}
