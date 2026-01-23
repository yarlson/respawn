package ui

import (
	"fmt"
	"time"
)

// Spinner provides a simple Braille-dot animation
type Spinner struct {
	frames []string
	index  int
}

// NewSpinner creates a new spinner with Braille dot pattern
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		index:  0,
	}
}

// Frame returns the current frame
func (s *Spinner) Frame() string {
	frame := s.frames[s.index%len(s.frames)]
	s.index++
	return frame
}

// Message returns a spinner message with the current frame
func (s *Spinner) Message(msg string) string {
	return fmt.Sprintf("%s %s", s.Frame(), msg)
}

// SpinnerTicker provides a time-based spinner
type SpinnerTicker struct {
	*Spinner
	ticker *time.Ticker
	done   chan bool
}

// NewSpinnerTicker creates a ticker-based spinner (for background updates)
func NewSpinnerTicker(interval time.Duration) *SpinnerTicker {
	return &SpinnerTicker{
		Spinner: NewSpinner(),
		ticker:  time.NewTicker(interval),
		done:    make(chan bool),
	}
}

// Stop stops the spinner ticker
func (st *SpinnerTicker) Stop() {
	st.ticker.Stop()
}

// C returns the channel for ticker events
func (st *SpinnerTicker) C() <-chan time.Time {
	return st.ticker.C
}
