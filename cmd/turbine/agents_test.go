package turbine

import (
	"bytes"
	"os"
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
	globalVerbose = false
	globalDebug = false

	cmd := RootCmd()
	cmd.SetArgs([]string{
		"agents",
		"--prd", "test.md",
		"--backend", "opencode",
		"--variant", "experimental",
		"--yes",
		"--verbose",
		"--debug",
	})

	// We expect an error since test.md doesn't exist, but flags should be parsed
	_ = cmd.Execute()

	assert.Equal(t, "test.md", agentsPrdPath)
	assert.Equal(t, "opencode", globalBackend)
	assert.Equal(t, "experimental", globalVariant)
	assert.True(t, globalYes)
	assert.True(t, globalVerbose)
	assert.True(t, globalDebug)
}
