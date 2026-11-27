package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/internal/utils"
	"github.com/mfridman/tparse/parse"
)

var usage = `tparse summarizes go test -json output. It shows failed tests first,
then a summary table for each package.

Usage:
    go test ./... -json | tparse [options...]
    go test ./... -json > test.json ; tparse -file test.json [options...]

Input:
    tparse reads go test -json output from a pipe or from -file. It does not
    read a redirected file (tparse < test.json). Use -file for saved output.
    In shell pipelines, use "set -o pipefail" so that the pipeline also fails
    when go test fails.

Output:
    Tables go to stdout. Errors, warnings, and build output go to stderr.
    Colors are on by default, also when stdout is not a terminal. For logs,
    files, and AI agents, use -nocolor or -format markdown.

Exit status:
    0 when all tests pass. Non-zero when a test fails, a package does not
    build, or tparse cannot read the input.

Options:
  Tables:
    -all                Show passed and skipped tests. Failed tests always show.
    -pass               Show a table of passed tests.
    -skip               Show a table of skipped tests.
    -slow N             With -pass or -all, show only the N slowest passed tests
                        in each package. Default is 0 (show all).
    -sort ATTR          Sort packages by name, elapsed, or cover.
                        Default is name.
    -notests            Show packages that have no test files.
    -smallscreen        Shorten package and test names to fit small screens.
    -trimpath PREFIX    Remove PREFIX from package names. Use "auto" to remove
                        the prefix that all packages share.

  Format:
    -format FORMAT      Table format: basic, plain, or markdown.
                        Default is basic.
    -nocolor            Disable all colors. NO_COLOR is also supported.

  Input and raw output:
    -file PATH          Read go test -json output from a file.
    -follow             Also print the raw go test output to stdout as it runs.
    -follow-output PATH Write the raw go test output to a file. Takes precedence
                        over -follow.
    -include-timestamp  Add a timestamp to each line of the raw output.
    -progress           Print one summary line for each package when it ends.
                        Useful for long test suites.
    -compare PATH       Compare against a previous go test -json output file.
                        (experimental)

  Other:
    -h                  Show help.
    -v                  Show version.

Examples:
    # Failures and summary only, with no ANSI colors (good for agents and logs)
    go test ./... -json | tparse -format markdown

    # GitHub Actions: show all tests and add them to the job summary
    set -o pipefail
    go test ./... -json | tparse -all -format markdown \
        | tee -a "$GITHUB_STEP_SUMMARY"

    # The 5 slowest passed tests in each package, slowest packages first
    go test ./... -json | tparse -pass -slow 5 -sort elapsed

    # Keep the full raw output in a file, and show only the summary
    go test ./... -json | tparse -follow-output test.log

    # Save the output once, then read it again with other options
    go test ./... -json > test.json ; tparse -file test.json -all
`

var version string

func main() {
	log.SetFlags(0)
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses the command line arguments, runs tparse, and returns the exit
// code. Invalid flag values return 2, the same exit code that the flag package
// uses for unknown flags.
func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("tparse", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprint(stderr, usage)
	}
	var (
		vPtr            = flags.Bool("v", false, "")
		versionPtr      = flags.Bool("version", false, "")
		hPtr            = flags.Bool("h", false, "")
		helpPtr         = flags.Bool("help", false, "")
		allPtr          = flags.Bool("all", false, "")
		passPtr         = flags.Bool("pass", false, "")
		skipPtr         = flags.Bool("skip", false, "")
		showNoTestsPtr  = flags.Bool("notests", false, "")
		smallScreenPtr  = flags.Bool("smallscreen", false, "")
		noColorPtr      = flags.Bool("nocolor", false, "")
		slowPtr         = flags.Int("slow", 0, "")
		fileNamePtr     = flags.String("file", "", "")
		formatPtr       = flags.String("format", "", "")
		followPtr       = flags.Bool("follow", false, "")
		followOutputPtr = flags.String("follow-output", "", "")
		sortPtr         = flags.String("sort", "name", "")
		progressPtr     = flags.Bool("progress", false, "")
		comparePtr      = flags.String("compare", "", "")
		trimPathPtr     = flags.String("trimpath", "", "")

		includeTimestamp = flags.Bool("include-timestamp", false, "")

		// Undocumented flags
		followVerbosePtr = flags.Bool("follow-verbose", false, "")

		// Legacy flags
		noBordersPtr = flags.Bool("noborders", false, "")
	)
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *vPtr || *versionPtr {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
		fmt.Fprintf(stdout, "tparse version: %s\n", version)
		return 0
	}
	if *hPtr || *helpPtr {
		fmt.Fprint(stdout, usage)
		return 0
	}
	format, err := parseFormat(*formatPtr, *noBordersPtr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	sorter, err := parseSorter(*sortPtr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if *allPtr {
		*passPtr = true
		*skipPtr = true
	}
	// Show colors by default.
	var disableColor bool
	if _, ok := os.LookupEnv("NO_COLOR"); ok || *noColorPtr {
		disableColor = true
	}

	var followOutput io.WriteCloser
	switch {
	case *followOutputPtr != "":
		followOutput, err = os.Create(*followOutputPtr)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		*followPtr = true
	case *followPtr, *followVerbosePtr:
		followOutput = utils.WriteNopCloser{Writer: stdout}
	default:
		// If no follow flags are set, we should not write to followOutput.
		followOutput = utils.WriteNopCloser{Writer: io.Discard}
	}
	// TODO(mf): we should marry the options with the flags to avoid having to do this.
	options := app.Options{
		Output:              stdout,
		DisableColor:        disableColor,
		FollowOutput:        *followPtr,
		FollowOutputWriter:  followOutput,
		FollowOutputVerbose: *followVerbosePtr,
		FileName:            *fileNamePtr,
		TestTableOptions: app.TestTableOptions{
			Pass:     *passPtr,
			Skip:     *skipPtr,
			Trim:     *smallScreenPtr,
			TrimPath: *trimPathPtr,
			Slow:     *slowPtr,
		},
		SummaryTableOptions: app.SummaryTableOptions{
			Trim:     *smallScreenPtr,
			TrimPath: *trimPathPtr,
		},
		Format:           format,
		Sorter:           sorter,
		ShowNoTests:      *showNoTestsPtr,
		Progress:         *progressPtr,
		ProgressOutput:   stdout,
		Compare:          *comparePtr,
		IncludeTimestamp: *includeTimestamp,

		// Do not expose publicly.
		DisableTableOutput: false,
	}
	exitCode, err := app.Run(options)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, parse.ErrNotParsable) {
			msg = "no parsable events: Make sure to run go test with -json flag"
		}
		fmt.Fprintln(stderr, msg)
	}
	return exitCode
}

func parseFormat(value string, noBorders bool) (app.OutputFormat, error) {
	switch value {
	case "basic":
		return app.OutputFormatBasic, nil
	case "plain":
		return app.OutputFormatPlain, nil
	case "markdown":
		return app.OutputFormatMarkdown, nil
	case "":
		// This was an existing flag, let's try to avoid breaking users.
		if noBorders {
			return app.OutputFormatPlain, nil
		}
		return app.OutputFormatBasic, nil
	default:
		return 0, fmt.Errorf("invalid option:%q. The -format flag must be one of: basic, plain or markdown", value)
	}
}

func parseSorter(value string) (parse.PackageSorter, error) {
	switch value {
	case "name":
		return parse.SortByPackageName, nil
	case "elapsed":
		return parse.SortByElapsed, nil
	case "cover":
		return parse.SortByCoverage, nil
	default:
		return nil, fmt.Errorf("invalid option:%q. The -sort flag must be one of: name, elapsed or cover", value)
	}
}
