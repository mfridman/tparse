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
	TestTableOptions    TestTableOptions
	SummaryTableOptions SummaryTableOptions

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

	cw := newConsoleWriter(w, option.Format, option.DisableColor)
	ec := make(chan parse.Event)
	defer close(ec)

	go displayEvent(cw, ec, option)

	summary, err := parse.Process(
		reader,
		parse.WithFollowOutput(option.FollowOutput),
		parse.WithWriter(w),
		parse.WithEvents(ec),
	)
	if err != nil {
		return 1, err
	}
	if len(summary.Packages) == 0 {
		return 1, fmt.Errorf("found no go test packages")
	}
	// Useful for tests that don't need additional output.
	if !option.DisableTableOutput {
		display(cw, summary, option)
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

func display(cw *consoleWriter, summary *parse.GoTestSummary, option Options) {
	// Sort packages by name ASC.
	packages := summary.GetSortedPackages(option.Sorter)
	// Only print the tests table if either pass or skip is true.
	if option.TestTableOptions.Pass || option.TestTableOptions.Skip {
		if option.Format == OutputFormatMarkdown {
			cw.testsTableMarkdown(packages, option.TestTableOptions)
		} else {
			cw.testsTable(packages, option.TestTableOptions)
		}
	}
	// Failures (if any) and summary table are always printed.
	cw.printFailed(packages)
	cw.summaryTable(packages, option.ShowNoTests, option.SummaryTableOptions)
}

// displayEvent reads from event channel ec and writes to cw when an event needs
// to be displayed.
func displayEvent(cw *consoleWriter, ec <-chan parse.Event, o Options) {
	for e := range ec {
		if cw.w == nil {
			continue
		}

		// Write plain text Output as if go test was run without the -json flag.
		if o.FollowOutput {
			if e.Output != "" {
				fmt.Fprint(cw.w, e.Output)
			}

			continue
		}

		// Display PASS lines for every successfully tested package.
		if e.LastLine() && e.Action == parse.ActionPass {
			fmt.Fprintln(cw.w, cw.styledLastLine(e))
		}
	}
}
