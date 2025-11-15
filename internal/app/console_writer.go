package app

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/mfridman/tparse/parse"
)

type OutputFormat int

const (
	// OutputFormatBasic is a normal table without a border
	OutputFormatPlain OutputFormat = iota + 1
	// OutputFormatBasic is a normal table with border
	OutputFormatBasic
	// OutputFormatBasic is a markdown-rendered table
	OutputFormatMarkdown
)

type consoleWriter struct {
	io.Writer
	format OutputFormat

	red    colorOptionFunc
	green  colorOptionFunc
	yellow colorOptionFunc
}

type colorOptionFunc func(s string) string

// newColor is a helper function to set the base color.
func newColor(color lipgloss.TerminalColor) colorOptionFunc {
	return func(text string) string {
		return lipgloss.NewStyle().Foreground(color).Render(text)
	}
}

// newMarkdownColor is a helper function to set the base color for markdown.
func newMarkdownColor(s string) colorOptionFunc {
	return func(text string) string {
		return s + " " + text
	}
}

func noColor() colorOptionFunc {
	return func(text string) string { return text }
}

func newConsoleWriter(w io.Writer, format OutputFormat, disableColor bool) *consoleWriter {
	if format == 0 {
		format = OutputFormatBasic
	}
	cw := &consoleWriter{
		Writer: w,
		format: format,
	}
	cw.red = noColor()
	cw.green = noColor()
	cw.yellow = noColor()

	if !disableColor {
		// NOTE(mf): GitHub Actions CI env (and probably others) do not have an
		// interactive TTY, and tparse through termenv will degrade to the
		// "best available option" .. which is no colors. We can work around this by
		// setting a color profile explicitly instead of relying on termenv to auto-detect.
		// Ref: https://github.com/charmbracelet/lipgloss/issues/74
		// Ref: https://github.com/mfridman/tparse/issues/76
		lipgloss.SetColorProfile(termenv.TrueColor)

		switch format {
		case OutputFormatMarkdown:
			cw.green = newMarkdownColor("🟢")
			cw.yellow = newMarkdownColor("🟡")
			cw.red = newMarkdownColor("🔴")
		default:
			cw.green = newColor(lipgloss.Color("10"))
			cw.yellow = newColor(lipgloss.Color("11"))
			cw.red = newColor(lipgloss.Color("9"))
		}
	}
	return cw
}

func (w *consoleWriter) FormatAction(action parse.Action) string {
	s := strings.ToUpper(action.String())
	switch action {
	case parse.ActionPass:
		return w.green(s)
	case parse.ActionSkip:
		return w.yellow(s)
	case parse.ActionFail:
		return w.red(s)
	default:
		return s
	}
}
