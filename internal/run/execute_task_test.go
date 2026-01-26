package run

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	relay "github.com/yarlson/relay"
	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/state"
	"github.com/yarlson/turbine/internal/tasks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	runFunc func(ctx context.Context, params relay.RunParams, events chan<- relay.Event) error
}

func (m *mockProvider) Name() string { return "mock" }

func (m *mockProvider) Run(ctx context.Context, params relay.RunParams, events chan<- relay.Event) error {
	_ = params
	defer close(events)
	if m.runFunc != nil {
		return m.runFunc(ctx, params, events)
	}
	return nil
}

func (m *mockProvider) Resume(ctx context.Context, sessionID string, params relay.RunParams, events chan<- relay.Event) error {
	_ = sessionID
	return m.Run(ctx, params, events)
}

func TestExecuteTask_Success(t *testing.T) {
	ctx := context.Background()
	repoDir := setupTestRepo(t)

	// Setup task.yaml
	tasksDir := filepath.Join(repoDir, ".turbine")
	require.NoError(t, os.MkdirAll(tasksDir, 0755))
	taskFile := &tasks.TaskFile{
		Version: 1,
		Task: tasks.Task{
			ID:            "T1",
			Title:         "Task 1",
			Status:        tasks.StatusTodo,
			Description:   "Description 1",
			CommitMessage: "feat: task 1",
			Verify:        []string{"true"},
		},
	}
	taskPath := filepath.Join(tasksDir, "task.yaml")
	require.NoError(t, taskFile.Save(taskPath))

	r := &Runner{
		RepoRoot: repoDir,
		TaskFile: taskFile,
		State:    &state.RunState{RunID: "test-run"},
		Config:   config.Defaults{Retry: config.Retry{Strokes: 3, Rotations: 3}},
	}

	mock := &mockProvider{}

	err := r.ExecuteTask(ctx, mock, "claude-3-5-sonnet", "")
	assert.NoError(t, err)

	// Verify task file removed and archived
	_, err = os.Stat(taskPath)
	assert.True(t, os.IsNotExist(err))
	archiveDir := filepath.Join(repoDir, ArchiveRelDir)
	entries, err := os.ReadDir(archiveDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)

	// Verify commit exists
	// We can't easily check the commit without shelling out, but setupTestRepo already uses git.
}

func TestExecuteTask_VerifyFailure(t *testing.T) {
	ctx := context.Background()
	repoDir := setupTestRepo(t)

	// Setup task.yaml
	tasksDir := filepath.Join(repoDir, ".turbine")
	require.NoError(t, os.MkdirAll(tasksDir, 0755))
	taskFile := &tasks.TaskFile{
		Version: 1,
		Task: tasks.Task{
			ID:            "T1",
			Title:         "Task 1",
			Status:        tasks.StatusTodo,
			Description:   "Description 1",
			CommitMessage: "feat: task 1",
			Verify:        []string{"false"}, // Fails
		},
	}
	taskPath := filepath.Join(tasksDir, "task.yaml")
	require.NoError(t, taskFile.Save(taskPath))

	r := &Runner{
		RepoRoot: repoDir,
		TaskFile: taskFile,
		State:    &state.RunState{RunID: "test-run"},
		Config:   config.Defaults{Retry: config.Retry{Strokes: 3, Rotations: 3}},
	}

	mock := &mockProvider{}

	err := r.ExecuteTask(ctx, mock, "claude-3-5-sonnet", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 3 rotations")

	// Verify status updated to failed (policy exhausted)
	updatedTasks, err := tasks.LoadTaskFile(taskPath)
	require.NoError(t, err)
	assert.Equal(t, tasks.StatusFailed, updatedTasks.Task.Status)
}
