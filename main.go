package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/parse"
)

// Flags.
var (
	vPtr           = flag.Bool("v", false, "")
	versionPtr     = flag.Bool("version", false, "")
	hPtr           = flag.Bool("h", false, "")
	helpPtr        = flag.Bool("help", false, "")
	allPtr         = flag.Bool("all", false, "")
	passPtr        = flag.Bool("pass", false, "")
	skipPtr        = flag.Bool("skip", false, "")
	showNoTestsPtr = flag.Bool("notests", false, "")
	smallScreenPtr = flag.Bool("smallscreen", false, "")
	noColorPtr     = flag.Bool("nocolor", false, "")
	slowPtr        = flag.Int("slow", 0, "")
	fileNamePtr    = flag.String("file", "", "")
	formatPtr      = flag.String("format", "", "")
	followPtr      = flag.Bool("follow", false, "")
	sortPtr        = flag.String("sort", "name", "")

	// TODO(mf): implement this
	ciPtr = flag.String("ci", "", "")

	// Legacy flags
	noBordersPtr = flag.Bool("noborders", false, "")
)

var usage = `Usage:
	go test ./... -json | tparse [options...]
	go test [packages...] -json | tparse [options...]
	go test [packages...] -json > pkgs.out ; tparse [options...] -file pkgs.out

Options:
	-h		Show help.
	-v		Show version.
	-all		Display table event for pass and skip. (Failed items always displayed)
	-pass		Display table for passed tests.
	-skip		Display table for skipped tests.
	-notests	Display packages containing no test files or empty test files.
	-smallscreen	Split subtest names vertically to fit on smaller screens.
	-slow		Number of slowest tests to display. Default is 0, display all.
	-sort           Sort table output by attribute [name, elapsed, cover]. Default is name.
	-nocolor	Disable all colors. (NO_COLOR also supported)
	-format		The output format for tables [basic, plain, markdown]. Default is basic.
	-file		Read test output from a file.
	-follow		Follow raw output as go test is running.
`

var (
	tparseVersion = ""
)

func main() {
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
	}
	flag.Parse()

	if *vPtr || *versionPtr {
		if buildInfo, ok := debug.ReadBuildInfo(); ok && buildInfo != nil && tparseVersion == "" {
			tparseVersion = buildInfo.Main.Version
		}
		fmt.Fprintf(os.Stdout, "tparse version: %s\n", tparseVersion)
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
		fmt.Fprintf(os.Stderr, "invalid option:%q. The -format flag must be one of: basic, plain or markdown", *formatPtr)
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
	options := app.Options{
		DisableColor: disableColor,
		FollowOutput: *followPtr,
		FileName:     *fileNamePtr,
		TestTableOptions: app.TestTableOptions{
			Pass: *passPtr,
			Skip: *skipPtr,
			Trim: *smallScreenPtr,
			Slow: *slowPtr,
		},
		Format:      format,
		Sorter:      sorter,
		ShowNoTests: *showNoTestsPtr,

		// Do not expose publically.
		DisableTableOutput: false,
	}
	exitCode, err := app.Run(os.Stdout, options)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, parse.ErrNotParseable) {
			msg = "no parseable events: Make sure to run go test with -json flag"
		}
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(exitCode)
}
