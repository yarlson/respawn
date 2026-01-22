package app

import (
	"fmt"
	"os"

	"github.com/yarlson/respawn/cmd/respawn"
)

// Execute executes the root command.
func Execute() {
	if err := respawn.RootCmd().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
