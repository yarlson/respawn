package gitx

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommitSavePoint creates a git commit with the given subject and footer.
// It includes all changes in the working tree (git add -A).
func CommitSavePoint(ctx context.Context, repoRoot, subjectLine, footerLine string) (string, error) {
	// 1. Stage all changes
	addCmd := exec.CommandContext(ctx, "git", "add", "-A")
	addCmd.Dir = repoRoot
	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}

	// 2. Create commit message
	commitMsg := fmt.Sprintf("%s\n\n%s", subjectLine, footerLine)

	// 3. Commit
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", commitMsg)
	commitCmd.Dir = repoRoot
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit: %w (output: %s)", err, string(output))
	}

	// 4. Get commit hash
	revCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	revCmd.Dir = repoRoot
	hashOutput, err := revCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}

	return strings.TrimSpace(string(hashOutput)), nil
}
