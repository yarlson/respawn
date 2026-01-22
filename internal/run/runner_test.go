package run

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"respawn/internal/tasks"

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

	// Create .gitignore to ignore .respawn by default in setup
	// so that its presence doesn't make the repo dirty unless we want it to.
	err = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".respawn/runs/\n.respawn/state/\n"), 0644)
	require.NoError(t, err)

	runCmd("git", "add", "README.md", ".gitignore")
	runCmd("git", "commit", "-m", "initial commit", "--no-gpg-sign")

	return dir
}

func TestRunner_Preflight(t *testing.T) {
	ctx := context.Background()

	t.Run("missing tasks.yaml", func(t *testing.T) {
		repoDir := setupTestRepo(t)
		_, err := NewRunner(ctx, Config{Cwd: repoDir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ".respawn/tasks.yaml missing")
	})

	t.Run("dirty tree error", func(t *testing.T) {
		repoDir := setupTestRepo(t)

		// Create .respawn/tasks.yaml
		err := os.MkdirAll(filepath.Join(repoDir, ".respawn"), 0755)
		require.NoError(t, err)
		taskList := tasks.TaskList{Version: 1, Tasks: []tasks.Task{{ID: "T1", Status: tasks.StatusTodo}}}
		err = taskList.Save(filepath.Join(repoDir, ".respawn", "tasks.yaml"))
		require.NoError(t, err)

		// Make it dirty with an UNTRACKED file that is NOT ignored
		err = os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("dirty"), 0644)
		require.NoError(t, err)

		_, err = NewRunner(ctx, Config{Cwd: repoDir})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "working tree is dirty")
	})

	t.Run("resume bypasses dirty check", func(t *testing.T) {
		repoDir := setupTestRepo(t)

		// Create .respawn/tasks.yaml
		err := os.MkdirAll(filepath.Join(repoDir, ".respawn", "state"), 0755)
		require.NoError(t, err)
		taskList := tasks.TaskList{Version: 1, Tasks: []tasks.Task{{ID: "T1", Status: tasks.StatusTodo}}}
		err = taskList.Save(filepath.Join(repoDir, ".respawn", "tasks.yaml"))
		require.NoError(t, err)

		// Create resume state
		err = os.WriteFile(filepath.Join(repoDir, ".respawn", "state", "run.json"), []byte(`{"run_id": "test"}`), 0644)
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

		err = os.MkdirAll(filepath.Join(repoDir, ".respawn"), 0755)
		require.NoError(t, err)
		taskList := tasks.TaskList{Version: 1, Tasks: []tasks.Task{{ID: "T1", Status: tasks.StatusTodo}}}
		err = taskList.Save(filepath.Join(repoDir, ".respawn", "tasks.yaml"))
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
		runCmd("git", "add", ".gitignore", ".respawn/tasks.yaml")
		runCmd("git", "commit", "-m", "prepare for ignore test", "--no-gpg-sign")

		// Preflight with auto-add
		_, err = NewRunner(ctx, Config{Cwd: repoDir, AutoAddIgnore: true})
		assert.NoError(t, err)

		// Verify .gitignore
		content, err := os.ReadFile(filepath.Join(repoDir, ".gitignore"))
		assert.NoError(t, err)
		assert.Contains(t, string(content), ".respawn/runs/")
		assert.Contains(t, string(content), ".respawn/state/")
	})
}

func TestRunner_PrintSummary(t *testing.T) {
	taskList := &tasks.TaskList{
		Tasks: []tasks.Task{
			{ID: "T1", Status: tasks.StatusDone},
			{ID: "T2", Status: tasks.StatusTodo},
			{ID: "T3", Status: tasks.StatusTodo, Deps: []string{"T2"}},
			{ID: "T4", Status: tasks.StatusFailed},
		},
	}
	runner := &Runner{Tasks: taskList}

	// This just verifies it doesn't panic and prints something
	// In a more thorough test we could capture stdout
	runner.PrintSummary()
}
