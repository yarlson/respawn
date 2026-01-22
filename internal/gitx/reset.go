package gitx

import (
	"context"
	"fmt"
	"os/exec"
)

// ResetHard performs a hard reset to the specified commit hash.
// It uses `git reset --hard <hash>` and clean untracked files that are likely to be artifacts.
func ResetHard(ctx context.Context, repoRoot, commitHash string) error {
	// 1. Reset hard
	resetCmd := exec.CommandContext(ctx, "git", "reset", "--hard", commitHash)
	resetCmd.Dir = repoRoot
	if out, err := resetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset hard to %s: %w (output: %s)", commitHash, err, string(out))
	}

	return nil
}
