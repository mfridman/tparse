package app

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mfridman/tparse/parse"
)

type Options struct {
	DisableTableOutput bool
	FollowOutput       bool
	DisableColor       bool
	Format             OutputFormat
	Sorter             parse.PackageSorter
	ShowNoTests        bool
	FileName           string

	// Test table options
	TestTableOptions TestTableOptions

	// TODO(mf): implement
	Progress bool
}

func Run(w io.Writer, option Options) (int, error) {
	var reader io.ReadCloser
	var err error
	if option.FileName != "" {
		if reader, err = os.Open(option.FileName); err != nil {
			return 1, err
		}
	} else {
		if reader, err = newPipeReader(); err != nil {
			return 1, errors.New("stdin must be a pipe, or use -file to open go test output file")
		}
	}
	defer reader.Close()

	summary, err := parse.Process(
		reader,
		parse.WithFollowOutput(option.FollowOutput),
		parse.WithWriter(w),
	)
	if err != nil {
		return 1, err
	}
	if len(summary.Packages) == 0 {
		return 1, fmt.Errorf("found no go test packages")
	}
	// Useful for tests that don't need additional output.
	if !option.DisableTableOutput {
		display(w, summary, option)
	}
	return summary.ExitCode(), nil
}

func newPipeReader() (io.ReadCloser, error) {
	finfo, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	// Check file mode bits to test for named pipe as stdin.
	if finfo.Mode()&os.ModeNamedPipe != 0 {
		return os.Stdin, nil
	}
	return nil, errors.New("stdin must be a pipe")
}

func display(w io.Writer, summary *parse.GoTestSummary, option Options) {
	cw := newConsoleWriter(w, option.Format, option.DisableColor)
	// Sort packages by name ASC.
	packages := summary.GetSortedPackages(option.Sorter)
	// Only print the tests table if either pass or skip is true.
	if option.TestTableOptions.Pass || option.TestTableOptions.Skip {
		cw.testsTable(packages, option.TestTableOptions)
	}
	// Failures (if any) and summary table are always printed.
	cw.printFailed(packages)
	cw.summaryTable(packages, option.ShowNoTests)
}
