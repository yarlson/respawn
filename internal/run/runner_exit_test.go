package run

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/state"
	"github.com/yarlson/turbine/internal/tasks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Run_ExitCodes(t *testing.T) {
	ctx := context.Background()

	t.Run("success all tasks", func(t *testing.T) {
		repoDir := setupTestRepo(t)
		tasksDir := filepath.Join(repoDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))

		taskList := &tasks.TaskList{
			Version: 1,
			Tasks: []tasks.Task{
				{
					ID:            "T1",
					Title:         "Task 1",
					Status:        tasks.StatusTodo,
					Verify:        []string{"true"},
					CommitMessage: "feat: task 1",
				},
				{
					ID:            "T2",
					Title:         "Task 2",
					Status:        tasks.StatusTodo,
					Deps:          []string{"T1"},
					Verify:        []string{"true"},
					CommitMessage: "feat: task 2",
				},
			},
		}
		tasksPath := filepath.Join(tasksDir, "tasks.yaml")
		require.NoError(t, taskList.Save(tasksPath))

		r := &Runner{
			RepoRoot: repoDir,
			Tasks:    taskList,
			State:    &state.RunState{RunID: "test-run"},
			Config:   config.Defaults{Retry: config.Retry{Strokes: 1, Rotations: 1}},
		}

		mock := &mockBackend{}
		err := r.Run(ctx, mock, "claude-3-5-sonnet")
		assert.NoError(t, err)

		// Verify state is cleared
		_, exists, err := state.Load(repoDir)
		assert.NoError(t, err)
		assert.False(t, exists)

		// Verify all tasks done
		updatedTasks, _ := tasks.Load(tasksPath)
		assert.Equal(t, tasks.StatusDone, updatedTasks.Tasks[0].Status)
		assert.Equal(t, tasks.StatusDone, updatedTasks.Tasks[1].Status)
	})

	t.Run("failure returns error", func(t *testing.T) {
		repoDir := setupTestRepo(t)
		tasksDir := filepath.Join(repoDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))

		taskList := &tasks.TaskList{
			Version: 1,
			Tasks: []tasks.Task{
				{
					ID:            "T1",
					Status:        tasks.StatusTodo,
					Verify:        []string{"false"},
					CommitMessage: "feat: task 1",
				},
			},
		}
		tasksPath := filepath.Join(tasksDir, "tasks.yaml")
		require.NoError(t, taskList.Save(tasksPath))

		r := &Runner{
			RepoRoot: repoDir,
			Tasks:    taskList,
			State:    &state.RunState{RunID: "test-run"},
			Config:   config.Defaults{Retry: config.Retry{Strokes: 1, Rotations: 1}},
		}

		mock := &mockBackend{}
		err := r.Run(ctx, mock, "claude-3-5-sonnet")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "1 tasks failed")

		// Verify state cleared anyway on normal completion (end of loop)
		_, exists, err := state.Load(repoDir)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
