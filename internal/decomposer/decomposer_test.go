package decomposer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"respawn/internal/backends"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	outputs []string
	err     error
	calls   int
}

func (m *mockBackend) StartSession(ctx context.Context, opts backends.SessionOptions) (string, error) {
	return "mock-session", nil
}

func (m *mockBackend) Send(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.calls >= len(m.outputs) {
		return nil, fmt.Errorf("too many calls to mockBackend")
	}
	out := m.outputs[m.calls]
	m.calls++
	return &backends.Result{Output: out}, nil
}

func TestDecompose(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := tmpDir
	prdPath := filepath.Join(tmpDir, "PRD.md")
	require.NoError(t, os.WriteFile(prdPath, []byte("Test PRD"), 0644))

	t.Run("successful decomposition", func(t *testing.T) {
		output := "```yaml\nversion: 1\ntasks:\n  - id: T-001\n    title: Task 1\n    status: todo\n    description: desc\n    commit_message: 'feat: t1'\n```"
		backend := &mockBackend{outputs: []string{output}}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "")
		require.NoError(t, err)

		tasksPath := filepath.Join(repoRoot, ".respawn", "tasks.yaml")
		assert.FileExists(t, tasksPath)
		content, err := os.ReadFile(tasksPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "id: T-001")
	})

	t.Run("successful retry", func(t *testing.T) {
		invalidOutput := "```yaml\ninvalid: yaml: :\n```"
		validOutput := "```yaml\nversion: 1\ntasks:\n  - id: T-001\n    title: Task 1\n    status: todo\n    description: desc\n    commit_message: 'feat: t1'\n```"
		backend := &mockBackend{outputs: []string{invalidOutput, validOutput}}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "")
		require.NoError(t, err)
		assert.Equal(t, 2, backend.calls)

		tasksPath := filepath.Join(repoRoot, ".respawn", "tasks.yaml")
		assert.FileExists(t, tasksPath)
	})

	t.Run("fail after max retries", func(t *testing.T) {
		invalidOutput := "```yaml\ninvalid: yaml: :\n```"
		backend := &mockBackend{outputs: []string{invalidOutput, invalidOutput, invalidOutput}}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decompose failed after 2 retries")
		assert.Equal(t, 3, backend.calls)
	})

	t.Run("missing PRD file", func(t *testing.T) {
		backend := &mockBackend{}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), "non-existent.md", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read PRD")
	})

	t.Run("no YAML in response", func(t *testing.T) {
		backend := &mockBackend{outputs: []string{"no yaml here", "no yaml here", "no yaml here"}}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), prdPath, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no YAML found")
	})
}
