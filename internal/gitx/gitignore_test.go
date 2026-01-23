package gitx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()

	runCmd := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Dir = repoRoot
		err := cmd.Run()
		require.NoError(t, err, "failed to run %s %v", name, args)
	}

	runCmd("git", "init")
	// Set user config for git operations in tests if needed
	runCmd("git", "config", "user.email", "test@example.com")
	runCmd("git", "config", "user.name", "Test User")

	return repoRoot
}

func TestMissingTurbineIgnores(t *testing.T) {
	ctx := context.Background()

	t.Run("none ignored", func(t *testing.T) {
		repoRoot := setupTestRepo(t)
		missing, err := MissingTurbineIgnores(ctx, repoRoot)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{".turbine/runs/", ".turbine/state/"}, missing)
	})

	t.Run("all ignored via .gitignore", func(t *testing.T) {
		repoRoot := setupTestRepo(t)
		err := os.WriteFile(filepath.Join(repoRoot, ".gitignore"), []byte(".turbine/runs/\n.turbine/state/\n"), 0644)
		require.NoError(t, err)

		missing, err := MissingTurbineIgnores(ctx, repoRoot)
		assert.NoError(t, err)
		assert.Empty(t, missing)
	})

	t.Run("partially ignored", func(t *testing.T) {
		repoRoot := setupTestRepo(t)
		err := os.WriteFile(filepath.Join(repoRoot, ".gitignore"), []byte(".turbine/runs/\n"), 0644)
		require.NoError(t, err)

		missing, err := MissingTurbineIgnores(ctx, repoRoot)
		assert.NoError(t, err)
		assert.Equal(t, []string{".turbine/state/"}, missing)
	})
}

func TestAddIgnoresToGitignore(t *testing.T) {
	t.Run("create new .gitignore", func(t *testing.T) {
		repoRoot := t.TempDir()
		ignores := []string{".turbine/runs/", ".turbine/state/"}

		err := AddIgnoresToGitignore(repoRoot, ignores)
		assert.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(repoRoot, ".gitignore"))
		assert.NoError(t, err)
		assert.Equal(t, ".turbine/runs/\n.turbine/state/\n", string(content))
	})

	t.Run("append to existing without trailing newline", func(t *testing.T) {
		repoRoot := t.TempDir()
		gitignorePath := filepath.Join(repoRoot, ".gitignore")
		err := os.WriteFile(gitignorePath, []byte("node_modules"), 0644)
		require.NoError(t, err)

		ignores := []string{".turbine/runs/"}
		err = AddIgnoresToGitignore(repoRoot, ignores)
		assert.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err)
		assert.Equal(t, "node_modules\n.turbine/runs/\n", string(content))
	})

	t.Run("append to existing with trailing newline", func(t *testing.T) {
		repoRoot := t.TempDir()
		gitignorePath := filepath.Join(repoRoot, ".gitignore")
		err := os.WriteFile(gitignorePath, []byte("node_modules\n"), 0644)
		require.NoError(t, err)

		ignores := []string{".turbine/runs/"}
		err = AddIgnoresToGitignore(repoRoot, ignores)
		assert.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err)
		assert.Equal(t, "node_modules\n.turbine/runs/\n", string(content))
	})

	t.Run("idempotency - no duplicates", func(t *testing.T) {
		repoRoot := t.TempDir()
		gitignorePath := filepath.Join(repoRoot, ".gitignore")
		err := os.WriteFile(gitignorePath, []byte(".turbine/runs/\n"), 0644)
		require.NoError(t, err)

		ignores := []string{".turbine/runs/", ".turbine/state/"}
		err = AddIgnoresToGitignore(repoRoot, ignores)
		assert.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err)
		assert.Equal(t, ".turbine/runs/\n.turbine/state/\n", string(content))
	})
}
