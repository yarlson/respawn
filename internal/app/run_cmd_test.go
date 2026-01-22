package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCmdFlags(t *testing.T) {
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
