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

	red    colorOptionFunc
	green  colorOptionFunc
	yellow colorOptionFunc
}

type colorOptionFunc func(s string, bold bool) string

// newColor is a helper function to set the base color.
func newColor(color lipgloss.TerminalColor) colorOptionFunc {
	return func(text string, bold bool) string {
		return lipgloss.NewStyle().Bold(bold).Foreground(color).Render(text)
	}
}

func noColor() colorOptionFunc {
	return func(text string, _ bool) string {
		return text
	}
}

func newConsoleWriter(w io.Writer, format OutputFormat, disableColor bool) *consoleWriter {
	if format == 0 {
		format = OutputFormatBasic
	}
	cw := &consoleWriter{
		w:      w,
		format: format,
	}
	if disableColor {
		cw.red = noColor()
		cw.green = noColor()
		cw.yellow = noColor()
	} else {
		// TODO(mf): not sure why I have to do this. It's working just fine locally but in
		// CI (GitHub Actions) it is not outputting with colors.
		// https://github.com/charmbracelet/lipgloss/issues/74
		// lipgloss.SetColorProfile(termenv.TrueColor)
		cw.red = newColor(lipgloss.Color("9"))
		cw.green = newColor(lipgloss.Color("10"))
		cw.yellow = newColor(lipgloss.Color("11"))
	}
	return cw
}
