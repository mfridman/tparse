package app

import (
	"io"

	"github.com/mfridman/tparse/parse"
)

func display(w io.Writer, packages parse.Packages, option Options) error {
	cw := newConsoleWriter(w, option.Format, option.DisableColor)

	cw.TestsTable(packages, testTableOptions{})
	cw.PrintFailed(packages)
	cw.SummaryTable(packages, option.ShowNoTests)

	return nil
}
