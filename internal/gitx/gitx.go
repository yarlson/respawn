package gitx

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// RepoRoot returns the absolute path to the root of the git repository
// containing the given directory.
func RepoRoot(ctx context.Context, cwd string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("detect repo root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsDirty returns true if the git repository at repoRoot has uncommitted changes.
func IsDirty(ctx context.Context, repoRoot string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("check dirty status: %w", err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}
