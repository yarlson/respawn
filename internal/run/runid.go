package run

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRunID creates a sortable unique run ID (timestamp + random suffix).
func GenerateRunID() string {
	ts := time.Now().Format("20060102-150405")
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 4)
	for i := range b {
		// Using math/rand (Go 1.22+ default source is automatically seeded)
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%s", ts, string(b))
}
