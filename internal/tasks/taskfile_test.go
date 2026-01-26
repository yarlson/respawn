package tasks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTaskFile(t *testing.T) {
	tmpDir := t.TempDir()
	taskPath := filepath.Join(tmpDir, "task.yaml")

	t.Run("valid task", func(t *testing.T) {
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
`
		err := os.WriteFile(taskPath, []byte(content), 0644)
		require.NoError(t, err)

		file, err := LoadTaskFile(taskPath)
		require.NoError(t, err)
		assert.Equal(t, 1, file.Version)
		assert.Equal(t, "T-001", file.Task.ID)
		assert.Equal(t, StatusTodo, file.Task.Status)
	})

	t.Run("invalid status", func(t *testing.T) {
		content := `
version: 1
task:
  id: T-001
  title: Task 1
  status: unknown
  description: desc 1
  commit_message: "feat: task 1"
`
		err := os.WriteFile(taskPath, []byte(content), 0644)
		require.NoError(t, err)

		_, err = LoadTaskFile(taskPath)
		assert.ErrorContains(t, err, "invalid status")
	})
}
