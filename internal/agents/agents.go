package agents

import (
	"context"
	"fmt"
	"os"

	"github.com/yarlson/turbine/internal/backends"
	"github.com/yarlson/turbine/internal/prompt"
)

type Generator struct {
	backend  backends.Backend
	repoRoot string
}

func New(backend backends.Backend, repoRoot string) *Generator {
	return &Generator{
		backend:  backend,
		repoRoot: repoRoot,
	}
}

// Generate instructs the coding agent to create AGENTS.md, supporting docs/,
// and CLAUDE.md symlink with progressive disclosure principles.
// The agent writes files directly using its tools (file write, mkdir, ln -s, etc.).
func (g *Generator) Generate(ctx context.Context, prdPath string, artifactsDir string, model, variant string) error {
	prdContent, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	sessionID, err := g.backend.StartSession(ctx, backends.SessionOptions{
		WorkingDir:   g.repoRoot,
		ArtifactsDir: artifactsDir,
		Model:        model,
		Variant:      variant,
	})
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}

	// Instruct the coding agent to generate and write all files
	userPrompt := prompt.AgentsSystemPrompt + "\n\n" + prompt.AgentsUserPrompt(string(prdContent))

	_, err = g.backend.Send(ctx, sessionID, userPrompt, backends.SendOptions{})
	if err != nil {
		return fmt.Errorf("send prompt: %w", err)
	}

	// Validate that required files were created
	if err := g.validateOutput(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// validateOutput checks that the agent created the required files.
func (g *Generator) validateOutput() error {
	// Check AGENTS.md exists
	agentsPath := g.repoRoot + "/AGENTS.md"
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md was not created")
	}

	// Check CLAUDE.md symlink exists
	claudePath := g.repoRoot + "/CLAUDE.md"
	info, err := os.Lstat(claudePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("CLAUDE.md symlink was not created")
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("CLAUDE.md exists but is not a symlink")
	}

	return nil
}
