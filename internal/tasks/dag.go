package tasks

// RunnableTasks returns a slice of tasks that are ready to be executed.
// A task is runnable if its status is StatusTodo and all its dependencies
// have StatusDone.
func RunnableTasks(tasks []Task) []Task {
	statusByID := make(map[string]TaskStatus)
	for _, t := range tasks {
		statusByID[t.ID] = t.Status
	}

	runnable := []Task{}
	for _, t := range tasks {
		if t.Status != StatusTodo {
			continue
		}

		allDepsDone := true
		for _, depID := range t.Deps {
			if statusByID[depID] != StatusDone {
				allDepsDone = false
				break
			}
		}

		if allDepsDone {
			runnable = append(runnable, t)
		}
	}

	return runnable
}

// BlockedSummary returns the count of tasks that are blocked.
// A task is blocked if it is StatusTodo but at least one of its
// dependencies is StatusFailed.
func BlockedSummary(tasks []Task) int {
	statusByID := make(map[string]TaskStatus)
	for _, t := range tasks {
		statusByID[t.ID] = t.Status
	}

	blockedCount := 0
	for _, t := range tasks {
		if t.Status != StatusTodo {
			continue
		}

		hasFailedDep := false
		for _, depID := range t.Deps {
			if statusByID[depID] == StatusFailed {
				hasFailedDep = true
				break
			}
		}

		if hasFailedDep {
			blockedCount++
		}
	}

	return blockedCount
}
