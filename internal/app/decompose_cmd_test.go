package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecomposeCmdFlags(t *testing.T) {
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
	cmd.SetArgs([]string{"decompose", "--prd", "test.md"})
	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test.md", prdPath)
}

func TestDecomposeCmdAllFlags(t *testing.T) {
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
	assert.NoError(t, err)

	assert.Equal(t, "test.md", prdPath)
	assert.Equal(t, "opencode", globalBackend)
	assert.Equal(t, "gpt-4", globalModel)
	assert.Equal(t, "experimental", globalVariant)
	assert.True(t, globalYes)
	assert.True(t, globalVerbose)
	assert.True(t, globalDebug)
}
