package app

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

type OutputFormat int

const (
	// OutputFormatBasic is a normal table withput a border
	OutputFormatPlain OutputFormat = iota + 1
	// OutputFormatBasic is a normal table with border
	OutputFormatBasic
	// OutputFormatBasic is a markdown-rendered table
	OutputFormatMarkdown
)

type consoleWriter struct {
	format OutputFormat
	w      io.Writer

	red    func(string, bool) string
	green  func(string, bool) string
	yellow func(string, bool) string
}

// newColor is a helper function to set the base color.
func newColor(color lipgloss.TerminalColor) func(text string, bold bool) string {
	return func(text string, bold bool) string {
		return lipgloss.NewStyle().Bold(bold).Foreground(color).Render(text)
	}
}

func newConsoleWriter(w io.Writer, format OutputFormat, disableColor bool) *consoleWriter {
	cw := &consoleWriter{
		w:      w,
		format: format,
	}
	if disableColor {
		cw.red = newColor(lipgloss.NoColor{})
		cw.green = newColor(lipgloss.NoColor{})
		cw.yellow = newColor(lipgloss.NoColor{})
	} else {
		cw.red = newColor(lipgloss.Color("9"))
		cw.green = newColor(lipgloss.Color("10"))
		cw.yellow = newColor(lipgloss.Color("11"))
	}
	return cw
}
