package tasks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TaskFile represents a single active task definition.
type TaskFile struct {
	Version int                    `yaml:"version"`
	Task    Task                   `yaml:"task"`
	Other   map[string]interface{} `yaml:",inline"`
}

// LoadTaskFile loads and validates a single task file.
func LoadTaskFile(path string) (*TaskFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task file: %w", err)
	}

	var tf TaskFile
	if err := yaml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("unmarshal task yaml: %w", err)
	}

	if err := tf.Validate(); err != nil {
		return nil, fmt.Errorf("validate task: %w", err)
	}

	return &tf, nil
}

// Save writes the task file to disk.
func (t *TaskFile) Save(path string) error {
	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshal task yaml: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write task file: %w", err)
	}

	return nil
}

// Validate ensures the task file is structurally valid.
func (t *TaskFile) Validate() error {
	if t.Version == 0 {
		return fmt.Errorf("version is required")
	}

	if t.Task.ID == "" {
		return fmt.Errorf("task ID is required")
	}

	if t.Task.Title == "" {
		return fmt.Errorf("task title is required")
	}

	if t.Task.Description == "" {
		return fmt.Errorf("task description is required")
	}

	switch t.Task.Status {
	case StatusTodo, StatusDone, StatusFailed:
		// ok
	default:
		return fmt.Errorf("invalid status \"%s\" for task %s (expected: todo, done, failed)", t.Task.Status, t.Task.ID)
	}

	if t.Task.Status != StatusDone && t.Task.CommitMessage == "" {
		return fmt.Errorf("task commit_message is required")
	}

	return nil
}
