package turbine

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentsCmdFlags(t *testing.T) {
	repoRoot := setupTestRepo(t)
	prdFile := repoRoot + "/test.md"
	require.NoError(t, os.WriteFile(prdFile, []byte("prd content"), 0644))

	// We need to reset flags for each test because they are global
	agentsPrdPath = ""

	cmd := RootCmd()

	// Test --prd is required
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"agents"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \"prd\" not set")

	// Test with --prd and invalid backend
	cmd.SetArgs([]string{"agents", "--prd", "test.md", "--backend", "nonexistent", "--yes"})
	err = cmd.Execute()
	assert.Error(t, err) // Fails due to invalid backend
	assert.Contains(t, err.Error(), "unknown backend")
}

func TestAgentsCmdFlagsCapture(t *testing.T) {
	// Reset flags
	agentsPrdPath = ""
	globalBackend = ""
	globalModel = ""
	globalVariant = ""
	globalYes = false

	cmd := RootCmd()
	cmd.SetArgs([]string{
		"agents",
		"--prd", "test.md",
		"--backend", "opencode",
		"--variant", "experimental",
		"--yes",
	})

	// We expect an error since test.md doesn't exist, but flags should be parsed
	_ = cmd.Execute()

	assert.Equal(t, "test.md", agentsPrdPath)
	assert.Equal(t, "opencode", globalBackend)
	assert.Equal(t, "experimental", globalVariant)
	assert.True(t, globalYes)
}

func setupTestRepo(t *testing.T) string {
	tmpDir := t.TempDir()
	// Initialize a dummy git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Create .turbine/runs to avoid errors
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".turbine", "runs"), 0755))

	// Move to tmpDir for the duration of the test
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldCwd))
	})

	return tmpDir
}
