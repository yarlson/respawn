package agents

import (
	"context"
	"fmt"
	"os"

	relay "github.com/yarlson/relay"
	filestore "github.com/yarlson/turbine/internal/relay/store"
	"github.com/yarlson/turbine/internal/relay/stream"
)

type Generator struct {
	backend  relay.Provider
	repoRoot string
}

func New(backend relay.Provider, repoRoot string) *Generator {
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

	// Instruct the coding agent to generate and write all files
	// AgentsGenerator is self-contained, no methodologies needed
	userPrompt := buildAgentsPrompt(string(prdContent))

	exec := relay.NewExecutor(g.backend)
	workflow := &relay.Workflow{
		WorkingDir: g.repoRoot,
		Sessions: []relay.Session{
			{
				Steps: []relay.Step{
					{
						Prompt:  userPrompt,
						Model:   model,
						Variant: variant,
						PostHook: func(_ *relay.StepContext, _ *relay.StepResult) error {
							if err := g.validateOutput(); err != nil {
								return fmt.Errorf("validation failed: %w", err)
							}
							return nil
						},
					},
				},
			},
		},
	}

	if err := g.runWorkflow(ctx, exec, workflow); err != nil {
		return err
	}

	_ = artifactsDir
	return nil
}

func (g *Generator) runWorkflow(ctx context.Context, exec *relay.Executor, workflow *relay.Workflow) error {
	events := make(chan relay.Event, 128)
	done := make(chan struct{})
	go func() {
		defer close(done)
		if workflow.ID == "" {
			for range events {
			}
			return
		}
		store := filestore.New(g.repoRoot)
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
