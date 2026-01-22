package run

import (
	"context"
	"os"
	"path/filepath"
	"respawn/internal/backends"
	"respawn/internal/state"
	"respawn/internal/tasks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	startSessionFunc func(ctx context.Context, opts backends.SessionOptions) (string, error)
	sendFunc         func(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error)
}

func (m *mockBackend) StartSession(ctx context.Context, opts backends.SessionOptions) (string, error) {
	if m.startSessionFunc != nil {
		return m.startSessionFunc(ctx, opts)
	}
	return "mock-session", nil
}

func (m *mockBackend) Send(ctx context.Context, sessionID string, prompt string, opts backends.SendOptions) (*backends.Result, error) {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, sessionID, prompt, opts)
	}
	return &backends.Result{Output: "mock-output"}, nil
}

func TestExecuteTask_Success(t *testing.T) {
	ctx := context.Background()
	repoDir := setupTestRepo(t)

	// Setup tasks.yaml
	tasksDir := filepath.Join(repoDir, ".respawn")
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
	}

	mock := &mockBackend{}

	err := r.ExecuteTask(ctx, mock)
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
	tasksDir := filepath.Join(repoDir, ".respawn")
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
	}

	mock := &mockBackend{}

	err := r.ExecuteTask(ctx, mock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")

	// Verify status NOT updated to done
	updatedTasks, err := tasks.Load(tasksPath)
	require.NoError(t, err)
	assert.Equal(t, tasks.StatusTodo, updatedTasks.Tasks[0].Status)
}
