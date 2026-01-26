package tasks

import "fmt"

// UpdateTaskStatus updates the status of a task by ID in the given task file.
func UpdateTaskStatus(path string, id string, status TaskStatus) error {
	tf, err := LoadTaskFile(path)
	if err != nil {
		return err
	}

	if tf.Task.ID != id {
		return fmt.Errorf("task not found: %s", id)
	}

	tf.Task.Status = status
	return tf.Save(path)
}
