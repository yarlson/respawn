package gitx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoRoot(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	// Outside of a git repo
	_, err := RepoRoot(ctx, tmp)
	assert.Error(t, err)

	// Inside a git repo
	runGit(t, tmp, "init")
	root, err := RepoRoot(ctx, tmp)
	assert.NoError(t, err)

	absTmp, _ := filepath.Abs(tmp)
	// On some systems (like macOS), TempDir might return a path with symlinks (e.g. /var -> /private/var)
	// git rev-parse --show-toplevel usually returns the evaluated real path.
	evalRoot, _ := filepath.EvalSymlinks(root)
	evalTmp, _ := filepath.EvalSymlinks(absTmp)
	assert.Equal(t, evalTmp, evalRoot)

	// In a subdirectory
	sub := filepath.Join(tmp, "sub")
	err = os.Mkdir(sub, 0755)
	assert.NoError(t, err)
	root, err = RepoRoot(ctx, sub)
	assert.NoError(t, err)
	evalRoot, _ = filepath.EvalSymlinks(root)
	assert.Equal(t, evalTmp, evalRoot)
}

func TestIsDirty(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	runGit(t, tmp, "init")

	// Clean repo
	dirty, err := IsDirty(ctx, tmp)
	assert.NoError(t, err)
	assert.False(t, dirty)

	// Dirty repo (untracked file)
	err = os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("hello"), 0644)
	assert.NoError(t, err)
	dirty, err = IsDirty(ctx, tmp)
	assert.NoError(t, err)
	assert.True(t, dirty)

	// Dirty repo (staged file)
	runGit(t, tmp, "add", "file.txt")
	dirty, err = IsDirty(ctx, tmp)
	assert.NoError(t, err)
	assert.True(t, dirty)

	// Clean repo (after commit)
	runGit(t, tmp, "config", "user.email", "you@example.com")
	runGit(t, tmp, "config", "user.name", "Your Name")
	runGit(t, tmp, "commit", "-m", "initial", "--no-gpg-sign")
	dirty, err = IsDirty(ctx, tmp)
	assert.NoError(t, err)
	assert.False(t, dirty)

	// Dirty repo (modified file)
	err = os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("world"), 0644)
	assert.NoError(t, err)
	dirty, err = IsDirty(ctx, tmp)
	assert.NoError(t, err)
	assert.True(t, dirty)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
}
