package app

import (
	"io"

	"github.com/mfridman/tparse/parse"
)

func display(w io.Writer, packages parse.Packages, option Options) error {
	cw := newConsoleWriter(w, option.Format, option.DisableColor)
	// Only display the tests table if either pass or skip is true.
	if option.TestTableOptions.Pass || option.TestTableOptions.Skip {
		cw.testsTable(packages, option.TestTableOptions)
	}
	// Always print failures (if any) and the summary table.
	cw.printFailed(packages)
	cw.summaryTable(packages, option.ShowNoTests)

	return nil
}
