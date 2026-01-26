package run

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	relay "github.com/yarlson/relay"

	"github.com/yarlson/turbine/internal/config"
	"github.com/yarlson/turbine/internal/state"
	"github.com/yarlson/turbine/internal/tasks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	dir := t.TempDir()

	runCmd := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test User",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test User",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "failed to run %s %v: %s", name, args, string(out))
	}

	runCmd("git", "init")

	// Need at least one commit for some git operations
	err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test Repo"), 0644)
	require.NoError(t, err)

	// Create .gitignore to ignore .turbine by default in setup
	// so that its presence doesn't make the repo dirty unless we want it to.
	err = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".turbine/runs/\n.turbine/state/\n"), 0644)
	require.NoError(t, err)

	runCmd("git", "add", "README.md", ".gitignore")
	runCmd("git", "commit", "-m", "initial commit", "--no-gpg-sign")

	return dir
}

func TestRunner_Preflight(t *testing.T) {
	ctx := context.Background()

	t.Run("missing prd file", func(t *testing.T) {
		repoDir := setupTestRepo(t)
		_, err := NewRunner(ctx, Config{Cwd: repoDir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PRD file not found")
	})

	t.Run("dirty tree error", func(t *testing.T) {
		repoDir := setupTestRepo(t)

		// Create .turbine/prd.md
		err := os.MkdirAll(filepath.Join(repoDir, ".turbine"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(repoDir, ".turbine", "prd.md"), []byte("Test PRD"), 0644)
		require.NoError(t, err)

		// Make it dirty with an UNTRACKED file that is NOT ignored
		err = os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("dirty"), 0644)
		require.NoError(t, err)

		_, err = NewRunner(ctx, Config{Cwd: repoDir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save your progress")
	})

	t.Run("resume bypasses dirty check", func(t *testing.T) {
		repoDir := setupTestRepo(t)

		// Create .turbine/prd.md and .turbine/task.yaml
		err := os.MkdirAll(filepath.Join(repoDir, ".turbine", "state"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(repoDir, ".turbine", "prd.md"), []byte("Test PRD"), 0644)
		require.NoError(t, err)
		taskFile := &tasks.TaskFile{
			Version: 1,
			Task:    tasks.Task{ID: "T1", Title: "Task 1", Status: tasks.StatusTodo, Description: "desc", CommitMessage: "feat: t1"},
		}
		err = taskFile.Save(filepath.Join(repoDir, ".turbine", "task.yaml"))
		require.NoError(t, err)

		// Create resume state
		err = os.WriteFile(filepath.Join(repoDir, ".turbine", "state", "run.json"), []byte(`{"run_id": "test", "active_task_id": "T1"}`), 0644)
		require.NoError(t, err)

		// Make it dirty
		err = os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("dirty"), 0644)
		require.NoError(t, err)

		runner, err := NewRunner(ctx, Config{Cwd: repoDir})
		assert.NoError(t, err)
		assert.True(t, runner.Resume)
	})

	t.Run("ignore-missing detection and auto-add", func(t *testing.T) {
		repoDir := setupTestRepo(t)

		// Create empty .gitignore to trigger missing detection
		err := os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte(""), 0644)
		require.NoError(t, err)

		err = os.MkdirAll(filepath.Join(repoDir, ".turbine"), 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(repoDir, ".turbine", "prd.md"), []byte("Test PRD"), 0644)
		require.NoError(t, err)

		// Commit them so the tree is clean for preflight
		runCmd := func(name string, args ...string) {
			cmd := exec.Command(name, args...)
			cmd.Dir = repoDir
			cmd.Env = append(os.Environ(),
				"GIT_AUTHOR_NAME=Test User",
				"GIT_AUTHOR_EMAIL=test@example.com",
				"GIT_COMMITTER_NAME=Test User",
				"GIT_COMMITTER_EMAIL=test@example.com",
			)
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "failed to run %s %v: %s", name, args, string(out))
		}
		runCmd("git", "add", ".gitignore", ".turbine/prd.md")
		runCmd("git", "commit", "-m", "prepare for ignore test", "--no-gpg-sign")

		// Preflight with auto-add
		_, err = NewRunner(ctx, Config{Cwd: repoDir, AutoAddIgnore: true})
		assert.NoError(t, err)

		// Verify .gitignore
		content, err := os.ReadFile(filepath.Join(repoDir, ".gitignore"))
		assert.NoError(t, err)
		assert.Contains(t, string(content), ".turbine/runs/")
		assert.Contains(t, string(content), ".turbine/state/")
	})
}

func TestRunner_PrintSummary(t *testing.T) {
	taskFile := &tasks.TaskFile{
		Version: 1,
		Task:    tasks.Task{ID: "T1", Title: "Task 1", Status: tasks.StatusTodo, Description: "desc", CommitMessage: "feat: t1"},
	}
	runner := &Runner{TaskFile: taskFile}

	// This just verifies it doesn't panic and prints something
	// In a more thorough test we could capture stdout
	runner.PrintSummary()
}

func TestRunner_Run_ExitCodes(t *testing.T) {
	ctx := context.Background()

	t.Run("success all tasks", func(t *testing.T) {
		repoDir := setupTestRepo(t)
		tasksDir := filepath.Join(repoDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))
		prdPath := filepath.Join(tasksDir, "prd.md")
		require.NoError(t, os.WriteFile(prdPath, []byte("Test PRD"), 0644))
		progressPath := filepath.Join(tasksDir, "progress.md")
		require.NoError(t, os.WriteFile(progressPath, []byte("# Progress\n"), 0644))

		taskFile := &tasks.TaskFile{
			Version: 1,
			Task: tasks.Task{
				ID:            "T1",
				Title:         "Task 1",
				Status:        tasks.StatusTodo,
				Description:   "Description 1",
				Verify:        []string{"true"},
				CommitMessage: "feat: task 1",
			},
		}
		taskPath := filepath.Join(tasksDir, "task.yaml")
		require.NoError(t, taskFile.Save(taskPath))

		r := &Runner{
			RepoRoot:     repoDir,
			TaskFile:     taskFile,
			State:        &state.RunState{RunID: "test-run"},
			Config:       config.Defaults{Retry: config.Retry{Strokes: 1, Rotations: 1}},
			PRDPath:      prdPath,
			ProgressPath: progressPath,
		}

		mock := &mockProvider{runFunc: func(_ context.Context, params relay.RunParams, _ chan<- relay.Event) error {
			taskPath := filepath.Join(params.WorkingDir, TaskRelPath)
			if _, err := os.Stat(taskPath); os.IsNotExist(err) {
				doneTask := &tasks.TaskFile{
					Version: 1,
					Task: tasks.Task{
						ID:          "T-DONE",
						Title:       "No remaining work",
						Status:      tasks.StatusDone,
						Description: "All PRD requirements satisfied.",
					},
				}
				return doneTask.Save(taskPath)
			}
			return nil
		}}
		err := r.Run(ctx, mock, Models{Fast: config.Model{Name: "claude-3-5-sonnet"}, Slow: config.Model{Name: "claude-4-5-opus"}})
		assert.NoError(t, err)

		// Verify state is cleared
		_, exists, err := state.Load(repoDir)
		assert.NoError(t, err)
		assert.False(t, exists)

		// Verify task file removed after success
		_, err = os.Stat(taskPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("failure returns error", func(t *testing.T) {
		repoDir := setupTestRepo(t)
		tasksDir := filepath.Join(repoDir, ".turbine")
		require.NoError(t, os.MkdirAll(tasksDir, 0755))

		taskFile := &tasks.TaskFile{
			Version: 1,
			Task: tasks.Task{
				ID:            "T1",
				Title:         "Task 1",
				Status:        tasks.StatusTodo,
				Description:   "Description 1",
				Verify:        []string{"false"},
				CommitMessage: "feat: task 1",
			},
		}
		taskPath := filepath.Join(tasksDir, "task.yaml")
		require.NoError(t, taskFile.Save(taskPath))

		r := &Runner{
			RepoRoot: repoDir,
			TaskFile: taskFile,
			State:    &state.RunState{RunID: "test-run"},
			Config:   config.Defaults{Retry: config.Retry{Strokes: 1, Rotations: 1}},
		}

		mock := &mockProvider{}
		err := r.Run(ctx, mock, Models{Fast: config.Model{Name: "claude-3-5-sonnet"}, Slow: config.Model{Name: "claude-4-5-opus"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed after 1 rotations")

		// Verify state kept for resume
		_, exists, err := state.Load(repoDir)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}
