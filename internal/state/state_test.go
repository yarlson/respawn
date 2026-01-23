package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateRoundtrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "turbine-state-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	expected := &RunState{
		RunID:               "run-123",
		ActiveTaskID:        "task-456",
		Rotation:            1,
		Stroke:              2,
		BackendName:         "opencode",
		BackendSessionID:    "sess-789",
		LastSavepointCommit: "abc123def",
		ArtifactRootPath:    "/tmp/artifacts",
	}

	// 1. Initial load should return exists=false
	state, exists, err := Load(tmpDir)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, state)

	// 2. Save state
	err = Save(tmpDir, expected)
	assert.NoError(t, err)

	// 3. Load state and verify
	actual, exists, err := Load(tmpDir)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, expected, actual)

	// 4. Verify file existence
	statePath := filepath.Join(tmpDir, ".turbine", "state", "run.json")
	assert.FileExists(t, statePath)

	// 5. Clear state
	err = Clear(tmpDir)
	assert.NoError(t, err)

	// 6. Verify file removed and load returns false
	assert.NoFileExists(t, statePath)
	state, exists, err = Load(tmpDir)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, state)
}

func TestClearNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "turbine-state-clear-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Clear on non-existent file should not error
	err = Clear(tmpDir)
	assert.NoError(t, err)
}
