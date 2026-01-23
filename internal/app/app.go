package app

import (
	"fmt"
	"os"

	"github.com/yarlson/turbine/cmd/turbine"
)

// Execute executes the root command.
func Execute() {
	if err := turbine.RootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
