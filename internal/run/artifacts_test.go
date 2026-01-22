package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewArtifacts(t *testing.T) {
	tempRoot := t.TempDir()
	runID := "test-run-123"

	artifacts, err := NewArtifacts(tempRoot, runID)
	require.NoError(t, err)
	assert.NotNil(t, artifacts)

	expectedRoot := filepath.Join(tempRoot, RunsDir, runID)
	assert.Equal(t, expectedRoot, artifacts.Root())

	// Verify directories exist
	subDirs := []string{
		SubDirPrompts,
		SubDirBackend,
		SubDirVerify,
		SubDirGit,
	}

	for _, sub := range subDirs {
		info, err := os.Stat(filepath.Join(expectedRoot, sub))
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	}
}

func TestArtifacts_WriteFile(t *testing.T) {
	tempRoot := t.TempDir()
	runID := "test-run-write"

	artifacts, err := NewArtifacts(tempRoot, runID)
	require.NoError(t, err)

	content := "hello world"
	filename := "test.txt"
	path, err := artifacts.WriteFile(SubDirPrompts, filename, content)
	require.NoError(t, err)

	expectedPath := filepath.Join(artifacts.Root(), SubDirPrompts, filename)
	assert.Equal(t, expectedPath, path)

	// Verify content
	got, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, content, string(got))
}

func TestArtifacts_WriteFile_InvalidSubDir(t *testing.T) {
	tempRoot := t.TempDir()
	runID := "test-run-invalid"

	artifacts, err := NewArtifacts(tempRoot, runID)
	require.NoError(t, err)

	_, err = artifacts.WriteFile("non-existent", "test.txt", "content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subdirectory non-existent does not exist")
}
