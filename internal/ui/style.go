package ui

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	resetCode = "\x1b[0m"

	// Foreground colors
	dimCode    = "\x1b[90m"
	whiteCode  = "\x1b[97m"
	cyanCode   = "\x1b[36m"
	greenCode  = "\x1b[32m"
	redCode    = "\x1b[31m"
	yellowCode = "\x1b[33m"
	boldCode   = "\x1b[1m"
	normalCode = "\x1b[22m"
)

// Style types
type Style int

const (
	StyleNormal Style = iota
	StyleBold
	StyleDim
	StyleCyan
	StyleGreen
	StyleRed
	StyleYellow
)

// styled applies ANSI styling to text
func styled(text string, s Style) string {
	if !isColorSupported() {
		return text
	}

	var code string
	switch s {
	case StyleBold:
		code = boldCode
	case StyleDim:
		code = dimCode
	case StyleCyan:
		code = cyanCode
	case StyleGreen:
		code = greenCode
	case StyleRed:
		code = redCode
	case StyleYellow:
		code = yellowCode
	default:
		code = whiteCode
	}

	return code + text + resetCode
}

// isColorSupported checks if colors should be used
func isColorSupported() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// Simple heuristic: colors on when terminal is interactive
	return true
}

// Styled text helpers
func Bold(text string) string {
	return styled(text, StyleBold)
}

func Dim(text string) string {
	return styled(text, StyleDim)
}

func Cyan(text string) string {
	return styled(text, StyleCyan)
}

func Green(text string) string {
	return styled(text, StyleGreen)
}

func Red(text string) string {
	return styled(text, StyleRed)
}

func Yellow(text string) string {
	return styled(text, StyleYellow)
}

// Divider returns a thin horizontal line
func Divider(width int) string {
	if width <= 0 {
		width = 40
	}
	return Dim(repeatChar("─", width))
}

// Section header with marker
func Section(marker, title string) string {
	return fmt.Sprintf("%s %s", styled(marker, StyleCyan), Bold(title))
}

// repeatChar creates a string by repeating a character
func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
