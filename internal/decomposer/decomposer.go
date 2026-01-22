package decomposer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"respawn/internal/backends"
	"respawn/internal/prompt"
	"respawn/internal/tasks"

	"gopkg.in/yaml.v3"
)

type Backend interface {
	backends.Backend
}

type Decomposer struct {
	Backend  Backend
	RepoRoot string
}

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

	res, err := d.Backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
	if err != nil {
		return fmt.Errorf("send prompt: %w", err)
	}

	tasksYAML := extractYAML(res.Output)
	if tasksYAML == "" {
		return fmt.Errorf("no YAML found in backend response")
	}

	var taskList tasks.TaskList
	if err := yaml.Unmarshal([]byte(tasksYAML), &taskList); err != nil {
		return fmt.Errorf("unmarshal tasks yaml: %w", err)
	}

	if err := taskList.Validate(); err != nil {
		return fmt.Errorf("validate tasks: %w", err)
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

func extractYAML(output string) string {
	// Try to find markdown code block first
	re := regexp.MustCompile("(?s)```(?:yaml)?\n(.*?)\n```")
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// If no code block, look for something that looks like YAML (version: 1 and tasks:)
	if strings.Contains(output, "version:") && strings.Contains(output, "tasks:") {
		return strings.TrimSpace(output)
	}

	return ""
}
