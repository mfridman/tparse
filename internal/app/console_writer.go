package app

import (
	"io"
	"os"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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
	disableColor bool
	format       OutputFormat
	w            io.Writer

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
		w:            w,
		format:       format,
		disableColor: disableColor,
	}
	cw.red = noColor()
	cw.green = noColor()
	cw.yellow = noColor()

	if !disableColor {
		// NOTE(mf): The GitHub Actions CI env (and probably others) does not have an
		// interactive TTY, and tparse will degrade to the "best available option" ..
		// which is no colors. We can work around this by setting the color profile
		// manually instead of relying on it to auto-detect.
		// Ref: https://github.com/charmbracelet/lipgloss/issues/74
		//
		// TODO(mf): Should this be an explicit env variable instead? Such as TPARSE_FORCE_COLOR
		//
		// For now we best-effort the most common CI environments and set this manually.
		if isCIEnvironment() {
			lipgloss.SetColorProfile(termenv.TrueColor)
		}
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

func isCIEnvironment() bool {
	if s := os.Getenv("CI"); s != "" {
		if ok, err := strconv.ParseBool(s); err == nil && ok {
			return true
		}
	}
	return false
}
