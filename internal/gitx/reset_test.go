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

func TestResetHard(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	runGit(t, tmp, "init")
	runGit(t, tmp, "config", "user.email", "test@example.com")
	runGit(t, tmp, "config", "user.name", "Test User")

	// Create initial commit
	file1 := filepath.Join(tmp, "file1.txt")
	err := os.WriteFile(file1, []byte("initial content"), 0644)
	require.NoError(t, err)
	runGit(t, tmp, "add", "file1.txt")
	runGit(t, tmp, "commit", "-m", "initial", "--no-gpg-sign")

	// Get initial commit hash
	initialHash := getHeadHash(t, tmp)

	// Modify tracked file and create another commit
	err = os.WriteFile(file1, []byte("modified content"), 0644)
	require.NoError(t, err)
	runGit(t, tmp, "add", "file1.txt")
	runGit(t, tmp, "commit", "-m", "second", "--no-gpg-sign")

	// Modify tracked file again but don't commit
	err = os.WriteFile(file1, []byte("uncommitted change"), 0644)
	require.NoError(t, err)

	// Create an untracked file
	untrackedFile := filepath.Join(tmp, "untracked.txt")
	err = os.WriteFile(untrackedFile, []byte("I should survive"), 0644)
	require.NoError(t, err)

	// Reset hard to initial commit
	err = ResetHard(ctx, tmp, initialHash)
	assert.NoError(t, err)

	// Verify tracked file is restored
	content, err := os.ReadFile(file1)
	assert.NoError(t, err)
	assert.Equal(t, "initial content", string(content))

	// Verify untracked file survives
	_, err = os.Stat(untrackedFile)
	assert.NoError(t, err, "untracked file should not be deleted")

	// Verify HEAD is at initial commit
	assert.Equal(t, initialHash, getHeadHash(t, tmp))
}

func getHeadHash(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	return strings.TrimSpace(string(out))
}
