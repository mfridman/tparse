package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func newTable(format OutputFormat) *table.Table {
	t := table.New()
	switch format {
	case OutputFormatPlain:
		t = t.Border(lipgloss.HiddenBorder()).BorderTop(false).BorderBottom(false)
	case OutputFormatMarkdown:
		t = t.Border(markdownBorder).BorderBottom(false).BorderTop(false)
	case OutputFormatBasic:
		t = t.Border(lipgloss.RoundedBorder())
	}
	return t
}

var markdownBorder = lipgloss.Border{
	Top:          "-",
	Bottom:       "-",
	Left:         "|",
	Right:        "|",
	TopLeft:      "", // empty for markdown
	TopRight:     "", // empty for markdown
	BottomLeft:   "", // empty for markdown
	BottomRight:  "", // empty for markdown
	MiddleLeft:   "|",
	MiddleRight:  "|",
	Middle:       "|",
	MiddleTop:    "|",
	MiddleBottom: "|",
}
