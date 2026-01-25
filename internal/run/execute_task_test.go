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

	// Setup tasks.yaml
	tasksDir := filepath.Join(repoDir, ".turbine")
	require.NoError(t, os.MkdirAll(tasksDir, 0755))
	taskList := &tasks.TaskList{
		Version: 1,
		Tasks: []tasks.Task{
			{
				ID:            "T1",
				Title:         "Task 1",
				Status:        tasks.StatusTodo,
				Description:   "Description 1",
				CommitMessage: "feat: task 1",
				Verify:        []string{"true"},
			},
		},
	}
	tasksPath := filepath.Join(tasksDir, "tasks.yaml")
	require.NoError(t, taskList.Save(tasksPath))

	r := &Runner{
		RepoRoot: repoDir,
		Tasks:    taskList,
		State:    &state.RunState{RunID: "test-run"},
		Config:   config.Defaults{Retry: config.Retry{Strokes: 3, Rotations: 3}},
	}

	mock := &mockProvider{}

	err := r.ExecuteTask(ctx, mock, "claude-3-5-sonnet", "")
	assert.NoError(t, err)

	// Verify status updated
	updatedTasks, err := tasks.Load(tasksPath)
	require.NoError(t, err)
	assert.Equal(t, tasks.StatusDone, updatedTasks.Tasks[0].Status)

	// Verify commit exists
	// We can't easily check the commit without shelling out, but setupTestRepo already uses git.
}

func TestExecuteTask_VerifyFailure(t *testing.T) {
	ctx := context.Background()
	repoDir := setupTestRepo(t)

	// Setup tasks.yaml
	tasksDir := filepath.Join(repoDir, ".turbine")
	require.NoError(t, os.MkdirAll(tasksDir, 0755))
	taskList := &tasks.TaskList{
		Version: 1,
		Tasks: []tasks.Task{
			{
				ID:            "T1",
				Title:         "Task 1",
				Status:        tasks.StatusTodo,
				Description:   "Description 1",
				CommitMessage: "feat: task 1",
				Verify:        []string{"false"}, // Fails
			},
		},
	}
	tasksPath := filepath.Join(tasksDir, "tasks.yaml")
	require.NoError(t, taskList.Save(tasksPath))

	r := &Runner{
		RepoRoot: repoDir,
		Tasks:    taskList,
		State:    &state.RunState{RunID: "test-run"},
		Config:   config.Defaults{Retry: config.Retry{Strokes: 3, Rotations: 3}},
	}

	mock := &mockProvider{}

	err := r.ExecuteTask(ctx, mock, "claude-3-5-sonnet", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 3 rotations")

	// Verify status updated to failed (policy exhausted)
	updatedTasks, err := tasks.Load(tasksPath)
	require.NoError(t, err)
	assert.Equal(t, tasks.StatusFailed, updatedTasks.Tasks[0].Status)
}
