package gitx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitSavePoint(t *testing.T) {
	// Create temp repo
	tmpDir, err := os.MkdirTemp("", "gitx-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v failed: %s", args, string(output))
	}

	runGit("init")
	// Set local config for test environment
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")
	runGit("config", "commit.gpgsign", "false")

	// Create a dummy file
	testFile := filepath.Join(tmpDir, "hello.txt")
	err = os.WriteFile(testFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	subject := "feat: initial commit"
	footer := "Respawn: T-001"

	hash, err := CommitSavePoint(ctx, tmpDir, subject, footer)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify hash matches HEAD
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tmpDir
	headOutput, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, hash, strings.TrimSpace(string(headOutput)))

	// Verify commit message
	cmd = exec.Command("git", "log", "-1", "--pretty=%B")
	cmd.Dir = tmpDir
	msgOutput, err := cmd.Output()
	require.NoError(t, err)

	expectedMsg := subject + "\n\n" + footer
	assert.Equal(t, expectedMsg, strings.TrimSpace(string(msgOutput)))
}
