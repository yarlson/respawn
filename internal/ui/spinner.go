package ui

import (
	"context"

	"github.com/yarlson/pin"
)

// Spinner wraps the pin library for consistent styling
type Spinner struct {
	p *pin.Pin
}

// NewSpinner creates a new styled spinner using pin library
func NewSpinner(message string) *Spinner {
	p := pin.New(message,
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorDefault),
		pin.WithPosition(pin.PositionLeft),
	)
	return &Spinner{p: p}
}

// Start begins the spinner animation
func (s *Spinner) Start(ctx context.Context) context.CancelFunc {
	return s.p.Start(ctx)
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(msg string) {
	s.p.UpdateMessage(msg)
}

// Stop ends the spinner with a success message
func (s *Spinner) Stop(msg string) {
	s.p.Stop(msg)
}

// Fail ends the spinner with a failure message
func (s *Spinner) Fail(msg string) {
	s.p.Fail(msg)
}
