package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func newTable(
	format OutputFormat,
	override func(style lipgloss.Style, row, col int) lipgloss.Style,
) *table.Table {
	tbl := table.New()
	switch format {
	case OutputFormatPlain:
		tbl.Border(lipgloss.HiddenBorder()).BorderTop(false).BorderBottom(false)
	case OutputFormatMarkdown:
		tbl.Border(markdownBorder).BorderBottom(false).BorderTop(false)
	case OutputFormatBasic:
		tbl.Border(lipgloss.RoundedBorder())
	}
	return tbl.StyleFunc(func(row, col int) lipgloss.Style {
		// Default style, may be overridden.
		style := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Align(lipgloss.Center)
		if override != nil {
			style = override(style, row, col)
		}
		return style
	})
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
