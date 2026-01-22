package gitx

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MissingRespawnIgnores checks if the required respawn directories are ignored by git.
// It returns a slice of paths that are NOT ignored.
func MissingRespawnIgnores(ctx context.Context, repoRoot string) ([]string, error) {
	required := []string{
		".respawn/runs/",
		".respawn/state/",
	}

	var missing []string
	for _, path := range required {
		// git check-ignore -q returns 0 if ignored, 1 if not ignored
		cmd := exec.CommandContext(ctx, "git", "check-ignore", "-q", path)
		cmd.Dir = repoRoot
		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				missing = append(missing, path)
				continue
			}
			return nil, fmt.Errorf("check-ignore for %s: %w", path, err)
		}
	}

	return missing, nil
}

// AddIgnoresToGitignore appends the missing ignores to the .gitignore file in repoRoot.
// It is idempotent and preserves the existing order and newlines.
func AddIgnoresToGitignore(repoRoot string, ignores []string) error {
	if len(ignores) == 0 {
		return nil
	}

	gitignorePath := filepath.Join(repoRoot, ".gitignore")

	// Check if file exists to decide whether to read it
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}

	existingLines := make(map[string]bool)
	if len(content) > 0 {
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				existingLines[line] = true
			}
		}
	}

	var toAdd []string
	for _, ignore := range ignores {
		if !existingLines[ignore] {
			toAdd = append(toAdd, ignore)
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore for append: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	// Ensure there's a newline before appending if the file wasn't empty and didn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("write newline to .gitignore: %w", err)
		}
	}

	for _, line := range toAdd {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write to .gitignore: %w", err)
		}
	}

	return nil
}
