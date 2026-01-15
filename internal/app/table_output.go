package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mfridman/tparse/parse"
)

type outputRow struct {
	testName    string
	packageName string
	output      string
}

func (r outputRow) toRow() []string {
	return []string{
		r.testName,
		r.packageName,
		r.output,
	}
}

func (c *consoleWriter) outputTable(packages []*parse.Package) {
	tbl := newTable(c.format, func(style lipgloss.Style, row, _ int) lipgloss.Style {
		switch row {
		case table.HeaderRow:
		default:
			style = style.Align(lipgloss.Left)
		}
		return style
	})
	header := outputRow{
		testName:    "Test",
		packageName: "Package",
		output:      "Output",
	}
	tbl.Headers(header.toRow()...)
	data := table.NewStringData()

	for _, pkg := range packages {
		printed := false
		for _, t := range pkg.Tests {
			if o := t.GetPrintableOutput(); o != "" {

				row := outputRow{
					packageName: t.Package,
					testName:    t.Name,
					output:      strings.TrimRight(o, "\n"),
				}
				if data.Rows() != 0 && !printed {
					// Add a blank row between packages.
					data.Append(outputRow{}.toRow())
					printed = true
				}
				data.Append(row.toRow())
			}
		}
	}

	if data.Rows() > 0 {
		fmt.Fprintln(c, tbl.Data(data).Render())
	}
}
