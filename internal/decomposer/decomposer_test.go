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

func TestPlanNext(t *testing.T) {
	validYAML := `version: 1
task:
  id: T-001
  title: Task 1
  status: todo
  description: desc
  commit_message: 'feat: t1'
`

	invalidYAML := `version: 1
task:
  id: T-001
  title: Task 1
  # missing status, description, commit_message
`

	// Helper to set up a fresh temp directory with PRD and progress files
	setupTempDir := func(t *testing.T) (string, string, string) {
		tmpDir := t.TempDir()
		prdPath := filepath.Join(tmpDir, "PRD.md")
		require.NoError(t, os.WriteFile(prdPath, []byte("Test PRD"), 0644))
		progressPath := filepath.Join(tmpDir, ".turbine", "progress.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(progressPath), 0755))
		require.NoError(t, os.WriteFile(progressPath, []byte("# Progress\n"), 0644))
		return tmpDir, prdPath, progressPath
	}

	// Helper to write task file
	writeTaskFile := func(repoRoot, content string) error {
		tasksDir := filepath.Join(repoRoot, ".turbine")
		if err := os.MkdirAll(tasksDir, 0755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(tasksDir, "task.yaml"), []byte(content), 0644)
	}

	// Note: With two-phase decomposition, call 0 is exploration, call 1+ is task generation
	opts := PlanOptions{
		FastModel: "fast-model",
		SlowModel: "slow-model",
	}

	t.Run("successful plan - backend writes valid file", func(t *testing.T) {
		repoRoot, prdPath, progressPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				// call 0 = exploration (no file written)
				// call 1 = task generation
				if call == 1 {
					return writeTaskFile(root, validYAML)
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.PlanNext(context.Background(), prdPath, progressPath, opts)
		require.NoError(t, err)
		assert.Equal(t, 2, backend.calls) // 1 explore + 1 generate

		taskPath := filepath.Join(repoRoot, ".turbine", "task.yaml")
		assert.FileExists(t, taskPath)
		content, err := os.ReadFile(taskPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "id: T-001")
	})

	t.Run("successful retry - backend fixes file on second attempt", func(t *testing.T) {
		repoRoot, prdPath, progressPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				// call 0 = exploration
				// call 1 = first generation attempt (invalid)
				// call 2 = retry (valid)
				if call == 1 {
					return writeTaskFile(root, invalidYAML)
				}
				if call >= 2 {
					return writeTaskFile(root, validYAML)
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.PlanNext(context.Background(), prdPath, progressPath, opts)
		require.NoError(t, err)
		assert.Equal(t, 3, backend.calls) // 1 explore + 2 generate attempts

		taskPath := filepath.Join(repoRoot, ".turbine", "task.yaml")
		assert.FileExists(t, taskPath)
	})

	t.Run("fail after max retries - backend keeps writing invalid file", func(t *testing.T) {
		repoRoot, prdPath, progressPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				if call >= 1 { // all generation attempts write invalid
					return writeTaskFile(root, invalidYAML)
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.PlanNext(context.Background(), prdPath, progressPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan failed after 2 retries")
		assert.Equal(t, 4, backend.calls) // 1 explore + 3 generate attempts
	})

	t.Run("fail - backend never creates file", func(t *testing.T) {
		repoRoot, prdPath, progressPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				// Backend doesn't write the file
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.PlanNext(context.Background(), prdPath, progressPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task file was not created")
	})

	t.Run("fail - backend writes empty file", func(t *testing.T) {
		repoRoot, prdPath, progressPath := setupTempDir(t)
		backend := &mockProvider{
			writeFile: func(root string, call int) error {
				if call >= 1 {
					return writeTaskFile(root, "")
				}
				return nil
			},
		}
		d := New(backend, repoRoot)

		err := d.PlanNext(context.Background(), prdPath, progressPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task file is empty")
	})

	t.Run("missing PRD file", func(t *testing.T) {
		repoRoot, _, progressPath := setupTempDir(t)
		backend := &mockProvider{}
		d := New(backend, repoRoot)
		err := d.PlanNext(context.Background(), "non-existent.md", progressPath, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read PRD")
	})
}

func TestValidateTaskFile(t *testing.T) {
	validYAML := `version: 1
task:
  id: T-001
  title: Task 1
  status: todo
  description: desc
  commit_message: 'feat: t1'
`

	t.Run("valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tasksDir := filepath.Join(tmpDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		taskPath := filepath.Join(tasksDir, "task.yaml")
		require.NoError(t, os.WriteFile(taskPath, []byte(validYAML), 0644))

		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTaskFile(taskPath)
		assert.NoError(t, err)
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTaskFile(filepath.Join(tmpDir, ".turbine", "task.yaml"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task file was not created")
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tasksDir := filepath.Join(tmpDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		taskPath := filepath.Join(tasksDir, "task.yaml")
		require.NoError(t, os.WriteFile(taskPath, []byte(""), 0644))

		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTaskFile(taskPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task file is empty")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		tasksDir := filepath.Join(tmpDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		taskPath := filepath.Join(tasksDir, "task.yaml")
		require.NoError(t, os.WriteFile(taskPath, []byte("invalid: yaml: :"), 0644))

		d := &Decomposer{repoRoot: tmpDir}
		err := d.validateTaskFile(taskPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse task")
	})
}
