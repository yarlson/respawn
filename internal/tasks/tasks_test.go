package tasks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.yaml")

	t.Run("valid tasks", func(t *testing.T) {
		content := `
version: 1
tasks:
  - id: T-001
    title: Task 1
    status: todo
    description: desc 1
    acceptance: ["a1"]
    verify: ["v1"]
    commit_message: "feat: task 1"
  - id: T-002
    title: Task 2
    status: todo
    deps: ["T-001"]
    description: desc 2
    acceptance: ["a2"]
    verify: ["v2"]
    commit_message: "feat: task 2"
`
		err := os.WriteFile(tasksPath, []byte(content), 0644)
		require.NoError(t, err)

		list, err := Load(tasksPath)
		require.NoError(t, err)
		assert.Equal(t, 1, list.Version)
		assert.Len(t, list.Tasks, 2)
		assert.Equal(t, "T-001", list.Tasks[0].ID)
		assert.Equal(t, StatusTodo, list.Tasks[0].Status)
		assert.Equal(t, []string{"T-001"}, list.Tasks[1].Deps)
	})

	t.Run("duplicate ids", func(t *testing.T) {
		content := `
version: 1
tasks:
  - id: T-001
    status: todo
  - id: T-001
    status: todo
`
		err := os.WriteFile(tasksPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = Load(tasksPath)
		assert.ErrorContains(t, err, "duplicate task ID: T-001")
	})

	t.Run("missing deps", func(t *testing.T) {
		content := `
version: 1
tasks:
  - id: T-001
    status: todo
    deps: ["T-002"]
`
		err := os.WriteFile(tasksPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = Load(tasksPath)
		assert.ErrorContains(t, err, "task T-001 depends on non-existent task: T-002")
	})

	t.Run("invalid status", func(t *testing.T) {
		content := `
version: 1
tasks:
  - id: T-001
    status: unknown
`
		err := os.WriteFile(tasksPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = Load(tasksPath)
		assert.ErrorContains(t, err, "invalid status for task T-001: unknown")
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := Load("non-existent.yaml")
		assert.Error(t, err)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		err := os.WriteFile(tasksPath, []byte("invalid: yaml: :"), 0644)
		require.NoError(t, err)

		_, err = Load(tasksPath)
		assert.Error(t, err)
	})
}
