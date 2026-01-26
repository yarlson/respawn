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
	tasksPath := filepath.Join(tmpDir, "task.yaml")

	content := `
version: 1
task:
  id: T-001
  title: Task 1
  status: todo
  description: desc 1
  acceptance: ["a1"]
  verify: ["v1"]
  commit_message: "feat: task 1"
  custom_field: "preserve me"
`
	err := os.WriteFile(tasksPath, []byte(content), 0644)
	require.NoError(t, err)

	t.Run("update to done", func(t *testing.T) {
		err := UpdateTaskStatus(tasksPath, "T-001", StatusDone)
		require.NoError(t, err)

		list, err := LoadTaskFile(tasksPath)
		require.NoError(t, err)
		assert.Equal(t, StatusDone, list.Task.Status)
		// Check if custom field is preserved
		data, _ := os.ReadFile(tasksPath)
		assert.Contains(t, string(data), "custom_field: preserve me")
	})

	t.Run("unknown id error", func(t *testing.T) {
		err := UpdateTaskStatus(tasksPath, "T-999", StatusDone)
		assert.ErrorContains(t, err, "task not found: T-999")
	})
}
