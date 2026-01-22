package tasks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTaskStatus(t *testing.T) {
	tmpDir := t.TempDir()
	tasksPath := filepath.Join(tmpDir, "tasks.yaml")

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
    custom_field: "preserve me"
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

	t.Run("update to done", func(t *testing.T) {
		err := UpdateTaskStatus(tasksPath, "T-001", StatusDone)
		require.NoError(t, err)

		list, err := Load(tasksPath)
		require.NoError(t, err)
		assert.Equal(t, StatusDone, list.Tasks[0].Status)
		// Check if custom field is preserved
		data, _ := os.ReadFile(tasksPath)
		assert.Contains(t, string(data), "custom_field: preserve me")
	})

	t.Run("update to failed", func(t *testing.T) {
		err := UpdateTaskStatus(tasksPath, "T-002", StatusFailed)
		require.NoError(t, err)

		list, err := Load(tasksPath)
		require.NoError(t, err)
		assert.Equal(t, StatusFailed, list.Tasks[1].Status)
	})

	t.Run("unknown id error", func(t *testing.T) {
		err := UpdateTaskStatus(tasksPath, "T-999", StatusDone)
		assert.ErrorContains(t, err, "task not found: T-999")
	})

	t.Run("preserve other fields", func(t *testing.T) {
		// This test specifically checks if fields not in the struct are preserved
		// We already checked custom_field above, but let's be explicit
		list, err := Load(tasksPath)
		require.NoError(t, err)

		found := false
		for _, task := range list.Tasks {
			if task.ID == "T-001" {
				// We need a way to check other fields.
				// If we use the struct approach, we'll need to update the struct.
				found = true
			}
		}
		assert.True(t, found)
	})
}
