package decomposer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	relay "github.com/yarlson/relay"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBackend simulates a coding agent that writes files directly.
type mockProvider struct {
	writeFile func(repoRoot string, call int) error
	repoRoot  string
	calls     int
	runErr    error
}

func (m *mockProvider) Name() string { return "mock" }

func (m *mockProvider) Run(ctx context.Context, params relay.RunParams, events chan<- relay.Event) error {
	_ = ctx
	defer close(events)
	if m.repoRoot == "" {
		m.repoRoot = params.WorkingDir
	}
	if m.runErr != nil {
		return m.runErr
	}
	if m.writeFile != nil {
		if err := m.writeFile(m.repoRoot, m.calls); err != nil {
			return err
		}
	}
	m.calls++
	return nil
}

func (m *mockProvider) Resume(ctx context.Context, sessionID string, params relay.RunParams, events chan<- relay.Event) error {
	_ = sessionID
	return m.Run(ctx, params, events)
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

	// Note: With two-phase decomposition, call 0 is exploration, call 1+ is task generation
	opts := DecomposeOptions{
		FastModel: "fast-model",
		SlowModel: "slow-model",
	}

	t.Run("successful decomposition - backend writes valid file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				// call 0 = exploration (no file written)
				// call 1 = task generation
				if call == 1 {
					return writeTasksFile(root, validYAML)
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, opts)
		require.NoError(t, err)
		assert.Equal(t, 2, backend.calls) // 1 explore + 1 generate

		tasksPath := filepath.Join(repoRoot, ".turbine", "tasks.yaml")
		assert.FileExists(t, tasksPath)
		content, err := os.ReadFile(tasksPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "id: T-001")
	})

	t.Run("successful retry - backend fixes file on second attempt", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				// call 0 = exploration
				// call 1 = first generation attempt (invalid)
				// call 2 = retry (valid)
				if call == 1 {
					return writeTasksFile(root, invalidYAML)
				}
				if call >= 2 {
					return writeTasksFile(root, validYAML)
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, opts)
		require.NoError(t, err)
		assert.Equal(t, 3, backend.calls) // 1 explore + 2 generate attempts

		tasksPath := filepath.Join(repoRoot, ".turbine", "tasks.yaml")
		assert.FileExists(t, tasksPath)
	})

	t.Run("fail after max retries - backend keeps writing invalid file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				if call >= 1 { // all generation attempts write invalid
					return writeTasksFile(root, invalidYAML)
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decompose failed after 2 retries")
		assert.Equal(t, 4, backend.calls) // 1 explore + 3 generate attempts
	})

	t.Run("fail - backend never creates file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				// Backend doesn't write the file
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tasks file was not created")
	})

	t.Run("fail - backend writes empty file", func(t *testing.T) {
		repoRoot, prdPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				if call >= 1 {
					return writeTasksFile(root, "")
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.Decompose(context.Background(), prdPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tasks file is empty")
	})

	t.Run("missing PRD file", func(t *testing.T) {
		repoRoot, _ := setupTempDir(t)
		backend := &mockProvider{}
		d := New(backend, repoRoot)
		err := d.Decompose(context.Background(), "non-existent.md", opts)
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
