package respawn

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRunTestRepo(t *testing.T) string {
	tmpDir := t.TempDir()

	runCmd := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test User",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test User",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		err := cmd.Run()
		require.NoError(t, err, "failed to run %s %v", name, args)
	}

	runCmd("git", "init")

	// Need at least one commit
	err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Repo"), 0644)
	require.NoError(t, err)

	// Create .respawn/tasks.yaml
	err = os.MkdirAll(filepath.Join(tmpDir, ".respawn"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, ".respawn", "tasks.yaml"), []byte("version: 1\ntasks: []"), 0644)
	require.NoError(t, err)

	// Create .gitignore to ignore .respawn stuff
	err = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(".respawn/runs/\n.respawn/state/\n"), 0644)
	require.NoError(t, err)

	runCmd("git", "add", ".")
	runCmd("git", "commit", "-m", "initial commit", "--no-gpg-sign")

	// Move to tmpDir for the duration of the test
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldCwd))
	})

	return tmpDir
}

func TestRunCmdFlags(t *testing.T) {
	setupRunTestRepo(t)

	// Reset global flags for clean state
	globalBackend = ""
	globalModel = ""
	globalVariant = ""
	globalYes = false
	globalVerbose = false
	globalDebug = false

	cmd := RootCmd()
	cmd.SetArgs([]string{
		"--backend", "claude",
		"--model", "claude-3-5-sonnet",
		"--variant", "fast",
		"--yes",
		"--verbose",
		"--debug",
	})

	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Equal(t, "claude", globalBackend)
	assert.Equal(t, "claude-3-5-sonnet", globalModel)
	assert.Equal(t, "fast", globalVariant)
	assert.True(t, globalYes)
	assert.True(t, globalVerbose)
	assert.True(t, globalDebug)
}

func TestRunCmdDefaults(t *testing.T) {
	setupRunTestRepo(t)

	// Reset global flags for clean state
	globalBackend = ""
	globalModel = ""
	globalVariant = ""
	globalYes = false
	globalVerbose = false
	globalDebug = false

	cmd := RootCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)

	assert.Equal(t, "", globalBackend)
	assert.Equal(t, "", globalModel)
	assert.Equal(t, "", globalVariant)
	assert.False(t, globalYes)
	assert.False(t, globalVerbose)
	assert.False(t, globalDebug)
}
