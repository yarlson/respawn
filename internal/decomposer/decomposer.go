package decomposer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"respawn/internal/backends"
	"respawn/internal/prompt"
	"respawn/internal/tasks"
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

	outputPath := ".respawn/tasks.yaml"

	sessionID, err := d.Backend.StartSession(ctx, backends.SessionOptions{
		WorkingDir:   d.RepoRoot,
		ArtifactsDir: artifactsDir,
	})
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	userPrompt := prompt.DecomposeUserPrompt(string(prdContent), outputPath)

	var lastRes *backends.Result
	var lastErr error
	var taskList *tasks.TaskList

	for i := 0; i <= maxValidationRetries; i++ {
		res, err := d.Backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
		if err != nil {
			return fmt.Errorf("send prompt: %w", err)
		}
		lastRes = res

		tasksYAML := extractYAML(res.Output)
		if tasksYAML == "" {
			lastErr = fmt.Errorf("no YAML found in backend response")
		} else {
			taskList, lastErr = validateTasksYAML(tasksYAML)
			if lastErr == nil {
				break
			}
		}

		if i < maxValidationRetries {
			userPrompt = prompt.DecomposeFixPrompt(string(prdContent), extractYAML(lastRes.Output), lastErr.Error())
		}
	}

	if lastErr != nil {
		return fmt.Errorf("decompose failed after %d retries: %w", maxValidationRetries, lastErr)
	}

	tasksPath := filepath.Join(d.RepoRoot, outputPath)
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0755); err != nil {
		return fmt.Errorf("create .respawn dir: %w", err)
	}

	if err := taskList.Save(tasksPath); err != nil {
		return fmt.Errorf("save tasks: %w", err)
	}

	return nil
}
