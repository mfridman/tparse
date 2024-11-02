package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mfridman/buildversion"
	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/internal/utils"
	"github.com/mfridman/tparse/parse"
)

// Flags.
var (
	vPtr            = flag.Bool("v", false, "")
	versionPtr      = flag.Bool("version", false, "")
	hPtr            = flag.Bool("h", false, "")
	helpPtr         = flag.Bool("help", false, "")
	allPtr          = flag.Bool("all", false, "")
	passPtr         = flag.Bool("pass", false, "")
	skipPtr         = flag.Bool("skip", false, "")
	showNoTestsPtr  = flag.Bool("notests", false, "")
	smallScreenPtr  = flag.Bool("smallscreen", false, "")
	noColorPtr      = flag.Bool("nocolor", false, "")
	slowPtr         = flag.Int("slow", 0, "")
	fileNamePtr     = flag.String("file", "", "")
	formatPtr       = flag.String("format", "", "")
	followPtr       = flag.Bool("follow", false, "")
	followOutputPtr = flag.String("follow-output", "", "")
	sortPtr         = flag.String("sort", "name", "")
	progressPtr     = flag.Bool("progress", false, "")
	comparePtr      = flag.String("compare", "", "")
	trimPathPtr     = flag.String("trimpath", "", "")
	// Undocumented flags
	followVerbosePtr = flag.Bool("follow-verbose", false, "")

	// Legacy flags
	noBordersPtr = flag.Bool("noborders", false, "")
)

var usage = `Usage:
    go test ./... -json | tparse [options...]
    go test [packages...] -json | tparse [options...]
    go test [packages...] -json > pkgs.out ; tparse [options...] -file pkgs.out

Options:
    -h             Show help.
    -v             Show version.
    -all           Display table event for pass and skip. (Failed items always displayed)
    -pass          Display table for passed tests.
    -skip          Display table for skipped tests.
    -notests       Display packages containing no test files or empty test files.
    -smallscreen   Split subtest names vertically to fit on smaller screens.
    -slow          Number of slowest tests to display. Default is 0, display all.
    -sort          Sort table output by attribute [name, elapsed, cover]. Default is name.
    -nocolor       Disable all colors. (NO_COLOR also supported)
    -format        The output format for tables [basic, plain, markdown]. Default is basic.
    -file          Read test output from a file.
    -follow        Follow raw output from go test to stdout.
    -follow-output Write raw output from go test to a file (takes precedence over -follow).
    -progress      Print a single summary line for each package. Useful for long running test suites.
    -compare       Compare against a previous test output file. (experimental)
    -trimpath      Remove path prefix from package names in output, simplifying their display.
`

var version string

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
	}
	flag.Parse()

	if *vPtr || *versionPtr {
		fmt.Fprintf(os.Stdout, "tparse version: %s\n", buildversion.New(version))
		return
	}
	if *hPtr || *helpPtr {
		fmt.Print(usage)
		return
	}
	var format app.OutputFormat
	switch *formatPtr {
	case "basic":
		format = app.OutputFormatBasic
	case "plain":
		format = app.OutputFormatPlain
	case "markdown":
		format = app.OutputFormatMarkdown
	case "":
		// This was an existing flag, let's try to avoid breaking users.
		format = app.OutputFormatBasic
		if *noBordersPtr {
			format = app.OutputFormatPlain
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid option:%q. The -format flag must be one of: basic, plain or markdown\n", *formatPtr)
		return
	}
	var sorter parse.PackageSorter
	switch *sortPtr {
	case "name":
		sorter = parse.SortByPackageName
	case "elapsed":
		sorter = parse.SortByElapsed
	case "cover":
		sorter = parse.SortByCoverage
	default:
		fmt.Fprintf(os.Stderr, "invalid option:%q. The -sort flag must be one of: name, elapsed or cover\n", *sortPtr)
		return
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
		var err error
		followOutput, err = os.Create(*followOutputPtr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		*followPtr = true
	case *followPtr:
		followOutput = os.Stdout
	default:
		// If no follow flags are set, we should not write to followOutput.
		followOutput = utils.WriteNopCloser{Writer: io.Discard}
	}
	// TODO(mf): we should marry the options with the flags to avoid having to do this.
	options := app.Options{
		Output:              os.Stdout,
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
		Format:         format,
		Sorter:         sorter,
		ShowNoTests:    *showNoTestsPtr,
		Progress:       *progressPtr,
		ProgressOutput: os.Stdout,
		Compare:        *comparePtr,

		// Do not expose publicly.
		DisableTableOutput: false,
	}
	exitCode, err := app.Run(options)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, parse.ErrNotParsable) {
			msg = "no parsable events: Make sure to run go test with -json flag"
		}
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(exitCode)
}
