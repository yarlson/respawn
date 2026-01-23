package turbine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	cmd := RootCmd()
	assert.Equal(t, "turbine", cmd.Use)
	assert.NotNil(t, cmd)
}
