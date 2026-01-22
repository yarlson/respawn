package run

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunVerification(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "respawn-verify-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("SuccessPath", func(t *testing.T) {
		artifacts, err := NewArtifacts(tmpDir, "test-run-success")
		require.NoError(t, err)

		commands := []string{
			"echo hello",
			"echo world",
		}
		results, err := RunVerification(context.Background(), artifacts, commands)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "echo hello", results[0].Command)
		assert.Equal(t, "echo world", results[1].Command)

		// Check logs
		content1, err := os.ReadFile(results[0].LogPath)
		assert.NoError(t, err)
		assert.Contains(t, string(content1), "hello")

		content2, err := os.ReadFile(results[1].LogPath)
		assert.NoError(t, err)
		assert.Contains(t, string(content2), "world")
	})

	t.Run("FailurePath", func(t *testing.T) {
		artifacts, err := NewArtifacts(tmpDir, "test-run-failure")
		require.NoError(t, err)

		commands := []string{
			"echo first",
			"false", // exits with code 1
			"echo third",
		}
		results, err := RunVerification(context.Background(), artifacts, commands)

		assert.Error(t, err)
		var vErr *VerifyError
		assert.ErrorAs(t, err, &vErr)
		assert.Equal(t, "false", vErr.Command)
		assert.Contains(t, vErr.LogPath, "02.log")

		assert.Len(t, results, 2) // Should stop after the failing command
		assert.Equal(t, "echo first", results[0].Command)
		assert.Equal(t, "false", results[1].Command)

		// Ensure 03.log was not created
		_, err = os.Stat(filepath.Join(artifacts.Root(), SubDirVerify, "03.log"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		artifacts, err := NewArtifacts(tmpDir, "test-run-cancel")
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		commands := []string{"sleep 10"}
		_, err = RunVerification(ctx, artifacts, commands)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
