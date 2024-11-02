package app

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mfridman/tparse/parse"
)

type Options struct {
	// Output is used to write the final output, such as the tables, summary, etc.
	Output io.Writer
	// DisableColor will disable all colors.
	DisableColor bool
	// Format will set the output format for tables.
	Format OutputFormat
	// Sorter will set the sort order for the table.
	Sorter parse.PackageSorter
	// ShowNoTests will display packages containing no test files or empty test files.
	ShowNoTests bool
	// FileName will read test output from a file.
	FileName string

	// Test table options
	TestTableOptions    TestTableOptions
	SummaryTableOptions SummaryTableOptions

	// FollowOutput will follow the raw output as go test is running.
	FollowOutput        bool           // Output to stdout
	FollowOutputWriter  io.WriteCloser // Output to a file, takes precedence over FollowOutput
	FollowOutputVerbose bool

	// Progress will print a single summary line for each package once the package has completed.
	// Useful for long running test suites. Maybe used with FollowOutput or on its own.
	//
	// This will output to stdout.
	Progress       bool
	ProgressOutput io.Writer

	// DisableTableOutput will disable all table output. This is used for testing.
	DisableTableOutput bool

	//
	//  Experimental
	//

	// Compare includes a diff of a previous test output file in the summary table.
	Compare string
}

func Run(option Options) (int, error) {
	var reader io.ReadCloser
	var err error
	if option.FileName != "" {
		if reader, err = os.Open(option.FileName); err != nil {
			return 1, err
		}
	} else {
		if reader, err = newPipeReader(); err != nil {
			return 1, errors.New("stdin must be a pipe, or use -file to open a go test output file")
		}
	}
	defer reader.Close()

	if option.FollowOutputWriter != nil {
		defer option.FollowOutputWriter.Close()
	}

	summary, err := parse.Process(
		reader,
		parse.WithFollowOutput(option.FollowOutput),
		parse.WithFollowVersboseOutput(option.FollowOutputVerbose),
		parse.WithWriter(option.FollowOutputWriter),
		parse.WithProgress(option.Progress),
		parse.WithProgressOutput(option.ProgressOutput),
	)
	if err != nil {
		return 1, err
	}
	if len(summary.Packages) == 0 {
		return 1, fmt.Errorf("found no go test packages")
	}
	// Useful for tests that don't need tparse table output. Very useful for testing output from
	// [parse.Process]
	if !option.DisableTableOutput {
		display(option.Output, summary, option)
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
	// Best effort to open the compare against file, if it exists.
	var warnings []string
	defer func() {
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "warning: %s\n", w)
		}
	}()
	var against *parse.GoTestSummary
	if option.Compare != "" {
		// TODO(mf): cleanup, this is messy.
		f, err := os.Open(option.Compare)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to open against file: %s", option.Compare))
		} else {
			defer f.Close()
			against, err = parse.Process(f)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to parse against file: %s", option.Compare))
			}
		}
	}

	cw := newConsoleWriter(w, option.Format, option.DisableColor)
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
	cw.summaryTable(packages, option.ShowNoTests, option.SummaryTableOptions, against)
}
