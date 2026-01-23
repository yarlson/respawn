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

func TestLoadCmdFlagsRequired(t *testing.T) {
	repoRoot := setupTestRepo(t)
	prdFile := filepath.Join(repoRoot, "test.md")
	require.NoError(t, os.WriteFile(prdFile, []byte("prd content"), 0644))

	prdPath = ""

	cmd := RootCmd()

	// Test --prd is required
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"load"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \"prd\" not set")
}
