package decomposer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yarlson/turbine/internal/backends"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBackend simulates a coding agent that writes files directly.
type mockBackend struct {
	// writeFile is called on each Send to simulate the backend writing a file
	writeFile func(repoRoot string, call int) error
	repoRoot  string
	calls     int
	sendErr   error
}

func (m *mockBackend) StartSession(ctx context.Context, opts backends.SessionOptions) (string, error) {
	m.repoRoot = opts.WorkingDir
	return "mock-session", nil
}

func (m *mockBackend) Send(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	if m.writeFile != nil {
		if err := m.writeFile(m.repoRoot, m.calls); err != nil {
			return nil, err
		}
	}
	m.calls++
	return &backends.Result{Output: ""}, nil
}

func TestDecompose(t *testing.T) {
	validYAML := `version: 1
tasks:
  - id: T-001
    title: Task 1
    status: todo
    description: desc
    commit_message: 'feat: t1'
`

	invalidYAML := `version: 1
tasks:
  - id: T-001
    title: Task 1
    # missing status, description, commit_message
`

	// Helper to set up a fresh temp directory with PRD file
	setupTempDir := func(t *testing.T) (string, string) {
		tmpDir := t.TempDir()
		prdPath := filepath.Join(tmpDir, "PRD.md")
		require.NoError(t, os.WriteFile(prdPath, []byte("Test PRD"), 0644))
		return tmpDir, prdPath
	}

	// Helper to write tasks file
	writeTasksFile := func(repoRoot, content string) error {
		tasksDir := filepath.Join(repoRoot, ".turbine")
		if err := os.MkdirAll(tasksDir, 0755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(tasksDir, "tasks.yaml"), []byte(content), 0644)
	}

	t.Run("successful decomposition - backend writes valid file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockBackend{
			writeFile: func(root string, call int) error {
				return writeTasksFile(root, validYAML)
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "", "claude-4-5-opus")
		require.NoError(t, err)
		assert.Equal(t, 1, backend.calls)

		tasksPath := filepath.Join(repoRoot, ".turbine", "tasks.yaml")
		assert.FileExists(t, tasksPath)
		content, err := os.ReadFile(tasksPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "id: T-001")
	})

	t.Run("successful retry - backend fixes file on second attempt", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockBackend{
			writeFile: func(root string, call int) error {
				if call == 0 {
					return writeTasksFile(root, invalidYAML)
				}
				return writeTasksFile(root, validYAML)
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "", "claude-4-5-opus")
		require.NoError(t, err)
		assert.Equal(t, 2, backend.calls)

		tasksPath := filepath.Join(repoRoot, ".turbine", "tasks.yaml")
		assert.FileExists(t, tasksPath)
	})

	t.Run("fail after max retries - backend keeps writing invalid file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockBackend{
			writeFile: func(root string, call int) error {
				return writeTasksFile(root, invalidYAML)
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "", "claude-4-5-opus")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decompose failed after 2 retries")
		assert.Equal(t, 3, backend.calls)
	})

	t.Run("fail - backend never creates file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockBackend{
			writeFile: func(root string, call int) error {
				// Backend doesn't write the file
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "", "claude-4-5-opus")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tasks file was not created")
	})

	t.Run("fail - backend writes empty file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockBackend{
			writeFile: func(root string, call int) error {
				return writeTasksFile(root, "")
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, "", "claude-4-5-opus")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tasks file is empty")
	})

	t.Run("missing PRD file", func(t *testing.T) {
		repoRoot, _ := setupTempDir(t)
		backend := &mockBackend{}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), "non-existent.md", "", "claude-4-5-opus")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read PRD")
	})
}

func TestValidateTasksFile(t *testing.T) {
	validYAML := `version: 1
tasks:
  - id: T-001
    title: Task 1
    status: todo
    description: desc
    commit_message: 'feat: t1'
`

	t.Run("valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tasksDir := filepath.Join(tmpDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		tasksPath := filepath.Join(tasksDir, "tasks.yaml")
		require.NoError(t, os.WriteFile(tasksPath, []byte(validYAML), 0644))

		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTasksFile(tasksPath)
		assert.NoError(t, err)
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTasksFile(filepath.Join(tmpDir, ".turbine", "tasks.yaml"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tasks file was not created")
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tasksDir := filepath.Join(tmpDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		tasksPath := filepath.Join(tasksDir, "tasks.yaml")
		require.NoError(t, os.WriteFile(tasksPath, []byte(""), 0644))

		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTasksFile(tasksPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tasks file is empty")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		tasksDir := filepath.Join(tmpDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		tasksPath := filepath.Join(tasksDir, "tasks.yaml")
		require.NoError(t, os.WriteFile(tasksPath, []byte("invalid: yaml: :"), 0644))

		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTasksFile(tasksPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse tasks")
	})
}
