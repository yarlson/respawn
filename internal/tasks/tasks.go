package tasks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type TaskStatus string

const (
	StatusTodo   TaskStatus = "todo"
	StatusDone   TaskStatus = "done"
	StatusFailed TaskStatus = "failed"
)

type Task struct {
	ID            string     `yaml:"id"`
	Title         string     `yaml:"title"`
	Status        TaskStatus `yaml:"status"`
	Deps          []string   `yaml:"deps,omitempty"`
	Description   string     `yaml:"description"`
	Acceptance    []string   `yaml:"acceptance"`
	Verify        []string   `yaml:"verify"`
	CommitMessage string     `yaml:"commit_message"`
}

type TaskList struct {
	Version int    `yaml:"version"`
	Tasks   []Task `yaml:"tasks"`
}

func Load(path string) (*TaskList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks file: %w", err)
	}

	var list TaskList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("unmarshal tasks yaml: %w", err)
	}

	if err := list.Validate(); err != nil {
		return nil, fmt.Errorf("validate tasks: %w", err)
	}

	return &list, nil
}

func (l *TaskList) Validate() error {
	ids := make(map[string]bool)
	for _, t := range l.Tasks {
		if t.ID == "" {
			return fmt.Errorf("task ID cannot be empty")
		}
		if ids[t.ID] {
			return fmt.Errorf("duplicate task ID: %s", t.ID)
		}
		ids[t.ID] = true

		switch t.Status {
		case StatusTodo, StatusDone, StatusFailed:
			// ok
		default:
			return fmt.Errorf("invalid status for task %s: %s", t.ID, t.Status)
		}
	}

	for _, t := range l.Tasks {
		for _, dep := range t.Deps {
			if !ids[dep] {
				return fmt.Errorf("task %s depends on non-existent task: %s", t.ID, dep)
			}
		}
	}

	return nil
}
