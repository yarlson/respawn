package decomposer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"respawn/internal/backends"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	output string
	err    error
}

func (m *mockBackend) StartSession(ctx context.Context, opts backends.SessionOptions) (string, error) {
	return "mock-session", nil
}

func (m *mockBackend) Send(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &backends.Result{Output: m.output}, nil
}

func TestDecompose(t *testing.T) {
	tmpDir := t.TempDir()
	repoRoot := tmpDir
	prdPath := filepath.Join(tmpDir, "PRD.md")
	require.NoError(t, os.WriteFile(prdPath, []byte("Test PRD"), 0644))

	t.Run("successful decomposition", func(t *testing.T) {
		output := "```yaml\nversion: 1\ntasks:\n  - id: T-001\n    title: Task 1\n    status: todo\n    description: desc\n    commit_message: 'feat: t1'\n```"
		backend := &mockBackend{output: output}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "")
		require.NoError(t, err)

		tasksPath := filepath.Join(repoRoot, ".respawn", "tasks.yaml")
		assert.FileExists(t, tasksPath)
		content, err := os.ReadFile(tasksPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "id: T-001")
	})

	t.Run("missing PRD file", func(t *testing.T) {
		backend := &mockBackend{}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), "non-existent.md", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read PRD")
	})

	t.Run("no YAML in response", func(t *testing.T) {
		backend := &mockBackend{output: "no yaml here"}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), prdPath, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no YAML found")
	})

	t.Run("invalid YAML response", func(t *testing.T) {
		backend := &mockBackend{output: "```yaml\ninvalid: yaml: :\n```"}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), prdPath, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal tasks yaml")
	})
}

func TestExtractYAML(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "with markdown fences",
			output:   "Here is your tasks file:\n```yaml\nversion: 1\ntasks: []\n```\nHope this helps!",
			expected: "version: 1\ntasks: []",
		},
		{
			name:     "plain yaml",
			output:   "version: 1\ntasks: []",
			expected: "version: 1\ntasks: []",
		},
		{
			name:     "no yaml",
			output:   "just some text",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractYAML(tt.output))
		})
	}
}
