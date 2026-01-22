package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	cmd := RootCmd()
	assert.Equal(t, "respawn", cmd.Use)
	assert.NotNil(t, cmd)
}
