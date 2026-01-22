package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRunID(t *testing.T) {
	id1 := GenerateRunID()
	id2 := GenerateRunID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "Run IDs should be unique")

	// Format is YYYYMMDD-HHMMSS-xxxx (15 + 1 + 4 = 20 chars)
	assert.Len(t, id1, 20, "Run ID should have expected length")

	// Check format with a simple check (could use regex but this is enough)
	assert.Contains(t, id1, "-")
}
