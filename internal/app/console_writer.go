package app

import (
	"image/color"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

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
func newColor(c color.Color) colorOptionFunc {
	return func(text string) string {
		return lipgloss.NewStyle().Foreground(c).Render(text)
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

	// NOTE(mf): lipgloss always renders full-color ANSI sequences, and leaves
	// color downsampling to the writer. We do not auto-detect the profile,
	// because GitHub Actions CI env (and probably others) do not have an
	// interactive TTY, and detection would degrade to no colors. So colors are
	// written as-is, unless the user disables them.
	// Ref: https://github.com/charmbracelet/lipgloss/issues/74
	// Ref: https://github.com/mfridman/tparse/issues/76
	if disableColor {
		cw.Writer = &colorprofile.Writer{Forward: w, Profile: colorprofile.NoTTY}
	} else {
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
