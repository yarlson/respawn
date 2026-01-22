package decomposer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractYAML(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "with markdown fences",
			output:   "Here is your tasks file:\n```yaml\nversion: 1\ntasks: []\n```\nHope this helps!",
			expected: "version: 1\ntasks: []",
		},
		{
			name:     "with yaml tag in fences",
			output:   "```yaml\nversion: 1\ntasks: []\n```",
			expected: "version: 1\ntasks: []",
		},
		{
			name:     "without yaml tag in fences",
			output:   "```\nversion: 1\ntasks: []\n```",
			expected: "version: 1\ntasks: []",
		},
		{
			name:     "plain yaml",
			output:   "version: 1\ntasks: []",
			expected: "version: 1\ntasks: []",
		},
		{
			name:     "no yaml",
			output:   "just some text",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractYAML(tt.output))
		})
	}
}

func TestValidateTasksYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		yaml := `
version: 1
tasks:
  - id: T-001
    title: Task 1
    status: todo
    description: desc
    commit_message: 'feat: t1'
`
		list, err := validateTasksYAML(yaml)
		assert.NoError(t, err)
		assert.NotNil(t, list)
		assert.Equal(t, "T-001", list.Tasks[0].ID)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		yaml := `invalid: yaml: :`
		list, err := validateTasksYAML(yaml)
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "unmarshal tasks yaml")
	})

	t.Run("invalid tasks schema", func(t *testing.T) {
		yaml := `
version: 1
tasks:
  - id: ""
    title: Task 1
`
		list, err := validateTasksYAML(yaml)
		assert.Error(t, err)
		assert.Nil(t, list)
		assert.Contains(t, err.Error(), "validate tasks")
	})
}
