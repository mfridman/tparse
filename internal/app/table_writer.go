package app

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

func newTableWriter(w io.Writer, format OutputFormat) *tablewriter.Table {
	tbl := tablewriter.NewWriter(w)
	tbl.SetAutoWrapText(false)
	switch format {
	case OutputFormatPlain:
		tbl.SetBorder(false)
		tbl.SetRowSeparator("")
		tbl.SetColumnSeparator("")
		tbl.SetHeaderLine(false)
	case OutputFormatMarkdown:
		tbl.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		tbl.SetCenterSeparator("|")
	case OutputFormatBasic:
		// TODO(mf): sigh, we're going to hack around this limitation by wrapping the table
		// with a lipgloss border around the un-borded tablewriter output. Wish upstream would
		// consider this PR:
		// https://github.com/olekukonko/tablewriter/pull/115
		tbl.SetBorder(false)
		tbl.SetRowSeparator("─")
		tbl.SetColumnSeparator("│")
		tbl.SetCenterSeparator("┼")
	}
	return tbl
}
