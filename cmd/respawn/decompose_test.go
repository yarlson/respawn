package respawn

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

	// Create .respawn/runs to avoid errors
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".respawn", "runs"), 0755))

	// Move to tmpDir for the duration of the test
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldCwd))
	})

	return tmpDir
}

func TestDecomposeCmdFlags(t *testing.T) {
	repoRoot := setupTestRepo(t)
	prdFile := filepath.Join(repoRoot, "test.md")
	require.NoError(t, os.WriteFile(prdFile, []byte("prd content"), 0644))

	// We need to reset flags for each test because they are global
	prdPath = ""

	cmd := RootCmd()

	// Test --prd is required
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"decompose"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s) \"prd\" not set")

	// Test with --prd
	cmd.SetArgs([]string{"decompose", "--prd", "test.md", "--yes"})
	err = cmd.Execute()
	assert.Error(t, err) // Still fails but now due to config/backend
	assert.Contains(t, err.Error(), "no YAML found")
}

func TestDecomposeCmdAllFlags(t *testing.T) {
	repoRoot := setupTestRepo(t)
	prdFile := filepath.Join(repoRoot, "test.md")
	require.NoError(t, os.WriteFile(prdFile, []byte("prd content"), 0644))

	// Reset flags
	prdPath = ""
	globalBackend = ""
	globalModel = ""
	globalVariant = ""
	globalYes = false
	globalVerbose = false
	globalDebug = false

	cmd := RootCmd()

	cmd.SetArgs([]string{
		"decompose",
		"--prd", "test.md",
		"--backend", "opencode",
		"--model", "gpt-4",
		"--variant", "experimental",
		"--yes",
		"--verbose",
		"--debug",
	})

	err := cmd.Execute()
	assert.Error(t, err) // Fails due to real backend call in test

	assert.Equal(t, "test.md", prdPath)
	assert.Equal(t, "opencode", globalBackend)
	assert.Equal(t, "gpt-4", globalModel)
	assert.Equal(t, "experimental", globalVariant)
	assert.True(t, globalYes)
	assert.True(t, globalVerbose)
	assert.True(t, globalDebug)
}
