package run

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yarlson/turbine/internal/tasks"
)

const (
	TaskRelPath     = ".turbine/task.yaml"
	ProgressRelPath = ".turbine/progress.md"
	ArchiveRelDir   = ".turbine/archive"
	PRDRelPath      = ".turbine/prd.md"
)

// EnsureProgressFile creates the progress file if it doesn't exist.
func EnsureProgressFile(repoRoot string) (string, error) {
	path := filepath.Join(repoRoot, ProgressRelPath)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat progress file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create progress dir: %w", err)
	}

	header := fmt.Sprintf("# Progress\n\n- %s Initialized progress log\n", time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(path, []byte(header), 0644); err != nil {
		return "", fmt.Errorf("write progress file: %w", err)
	}

	return path, nil
}

// AppendProgress appends a single entry to the progress log.
func AppendProgress(repoRoot, entry string) error {
	path, err := EnsureProgressFile(repoRoot)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read progress file: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open progress file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if len(data) > 0 && data[len(data)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("write progress newline: %w", err)
		}
	}

	if _, err := f.WriteString(entry + "\n"); err != nil {
		return fmt.Errorf("append progress entry: %w", err)
	}

	return nil
}

// ArchiveTaskFile saves a copy of the task file into the archive directory.
func ArchiveTaskFile(repoRoot string, taskFile *tasks.TaskFile) (string, error) {
	if taskFile == nil {
		return "", fmt.Errorf("task file is nil")
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, ArchiveRelDir), 0755); err != nil {
		return "", fmt.Errorf("create archive dir: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	filename := fmt.Sprintf("%s-%s.yaml", timestamp, taskFile.Task.ID)
	path := filepath.Join(repoRoot, ArchiveRelDir, filename)

	if err := taskFile.Save(path); err != nil {
		return "", fmt.Errorf("archive task: %w", err)
	}

	return path, nil
}
