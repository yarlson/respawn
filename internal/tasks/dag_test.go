package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnableTasks(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []Task
		expected []string
	}{
		{
			name: "single runnable task",
			tasks: []Task{
				{ID: "T1", Status: StatusTodo},
			},
			expected: []string{"T1"},
		},
		{
			name: "multiple runnable tasks",
			tasks: []Task{
				{ID: "T1", Status: StatusTodo},
				{ID: "T2", Status: StatusTodo},
			},
			expected: []string{"T1", "T2"},
		},
		{
			name: "task with dependencies done",
			tasks: []Task{
				{ID: "T1", Status: StatusDone},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
			},
			expected: []string{"T2"},
		},
		{
			name: "task with dependencies todo",
			tasks: []Task{
				{ID: "T1", Status: StatusTodo},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
			},
			expected: []string{"T1"},
		},
		{
			name: "task with dependencies failed",
			tasks: []Task{
				{ID: "T1", Status: StatusFailed},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
			},
			expected: []string{},
		},
		{
			name: "stable ordering preserved",
			tasks: []Task{
				{ID: "C", Status: StatusTodo},
				{ID: "A", Status: StatusTodo},
				{ID: "B", Status: StatusTodo},
			},
			expected: []string{"C", "A", "B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runnable := RunnableTasks(tt.tasks)
			ids := []string{}
			for _, rt := range runnable {
				ids = append(ids, rt.ID)
			}
			assert.Equal(t, tt.expected, ids)
		})
	}
}

func TestBlockedSummary(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []Task
		expected int
	}{
		{
			name: "no blocked tasks",
			tasks: []Task{
				{ID: "T1", Status: StatusTodo},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
			},
			expected: 0,
		},
		{
			name: "one blocked task",
			tasks: []Task{
				{ID: "T1", Status: StatusFailed},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
			},
			expected: 1,
		},
		{
			name: "multiple blocked tasks",
			tasks: []Task{
				{ID: "T1", Status: StatusFailed},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
				{ID: "T3", Status: StatusTodo, Deps: []string{"T1"}},
			},
			expected: 2,
		},
		{
			name: "blocked by transitive failure",
			tasks: []Task{
				{ID: "T1", Status: StatusFailed},
				{ID: "T2", Status: StatusTodo, Deps: []string{"T1"}},
				{ID: "T3", Status: StatusTodo, Deps: []string{"T2"}},
			},
			// T2 is blocked by T1. T3 is NOT directly blocked by a failed dep.
			// PRD says: "blocked due to failed deps".
			// Let's re-read: "A task is blocked if it is StatusTodo but at least one of its dependencies is StatusFailed."
			// This usually refers to direct dependencies.
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, BlockedSummary(tt.tasks))
		})
	}
}
