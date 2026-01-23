package ui

import "fmt"

// TaskStatus represents the visual status of a task
type TaskStatus string

const (
	StatusDone       TaskStatus = "✓"
	StatusFailed     TaskStatus = "✗"
	StatusInProgress TaskStatus = "◐"
	StatusBlocked    TaskStatus = "○"
	StatusPending    TaskStatus = "·"
)

// TaskStatusColor returns the styled status symbol
func TaskStatusColor(status TaskStatus) string {
	switch status {
	case StatusDone:
		return Green(string(status))
	case StatusFailed:
		return Red(string(status))
	case StatusInProgress:
		return Cyan(string(status))
	case StatusBlocked:
		return Yellow(string(status))
	default:
		return Dim(string(status))
	}
}

// TaskLine formats a task with its status
func TaskLine(status TaskStatus, id, title string) string {
	statusStr := TaskStatusColor(status)
	idStr := Dim(id)
	return fmt.Sprintf("%s %s %s", statusStr, idStr, title)
}

// TaskEntry formats a task for listing (without status coloring)
func TaskEntry(id, title string) string {
	return fmt.Sprintf("· %s %s", Dim(id), title)
}

// SuccessMarker returns a styled success indicator
func SuccessMarker() string {
	return Green("✓")
}

// FailureMarker returns a styled failure indicator
func FailureMarker() string {
	return Red("✗")
}

// InProgressMarker returns a styled in-progress indicator
func InProgressMarker() string {
	return Cyan("◐")
}

// SummaryStats formats task counts into a summary line
func SummaryStats(total, done, ready, blocked, failed int) string {
	return fmt.Sprintf(
		"%d tasks: %s %s %s %s",
		total,
		Green(fmt.Sprintf("%d done", done)),
		Cyan(fmt.Sprintf("%d ready", ready)),
		Yellow(fmt.Sprintf("%d blocked", blocked)),
		Red(fmt.Sprintf("%d failed", failed)),
	)
}
