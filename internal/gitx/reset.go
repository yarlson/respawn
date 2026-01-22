package gitx

import (
	"context"
	"fmt"
	"os/exec"
)

// ResetHard performs a hard reset to the specified commit hash.
// It uses `git reset --hard <hash>` and does not clean untracked files.
func ResetHard(ctx context.Context, repoRoot, commitHash string) error {
	cmd := exec.CommandContext(ctx, "git", "reset", "--hard", commitHash)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset hard to %s: %w (output: %s)", commitHash, err, string(out))
	}
	return nil
}
