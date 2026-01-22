package tasks

import "fmt"

// UpdateTaskStatus updates the status of a task by ID in the given tasks file.
func UpdateTaskStatus(path string, id string, status TaskStatus) error {
	list, err := Load(path)
	if err != nil {
		return err
	}

	found := false
	for i := range list.Tasks {
		if list.Tasks[i].ID == id {
			list.Tasks[i].Status = status
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found: %s", id)
	}

	return list.Save(path)
}
