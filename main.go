package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mfridman/tparse/parse"
	"github.com/mfridman/tparse/version"

	colorable "github.com/mattn/go-colorable"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// Flags.
var (
	versionPtr     = flag.Bool("v", false, "")
	allPtr         = flag.Bool("all", false, "")
	passPtr        = flag.Bool("pass", false, "")
	skipPtr        = flag.Bool("skip", false, "")
	showNoTestsPtr = flag.Bool("notests", false, "")
	dumpPtr        = flag.Bool("dump", false, "") // TODO(mf): rename this to -replay with v1
	smallScreenPtr = flag.Bool("smallscreen", false, "")
	topPtr         = flag.Bool("top", false, "") // TODO(mf): rename this to -reverse with v1
	noColorPtr     = flag.Bool("nocolor", false, "")
)

var usage = `Usage:
	go test ./... -json | tparse [options...]
	go test [packages...] -json | tparse [options...]
	go test [packages...] -json > pkgs.out ; tparse [options...] pkgs.out

Options:
	-h		Show help.
	-v		Show version.
	-all		Display table event for pass and skip. (Failed items displayed regardless)
	-pass		Display table for passed tests.
	-skip		Display table for skipped tests.
	-notests	Display packages containing no test files or empty test files in summary.
	-dump		Enables recovering go test output in non-JSON format.
	-smallscreen	Split subtest names vertically to fit on smaller screens.
	-top		Display summary table towards top.
	-nocolor	Disable all colors.
`

type consoleWriter struct {
	Color  bool
	Output io.Writer
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprint(usage))
		os.Exit(2)
	}
	flag.Parse()

	if *versionPtr {
		fmt.Fprintf(os.Stdout, "tparse version: %s\n", version.Version())
		os.Exit(0)
	}

	r, err := newReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flag.Usage()
	}
	defer r.Close()

	var replayBuf bytes.Buffer
	tr := io.TeeReader(r, &replayBuf)

	pkgs, err := parse.Process(tr)
	if err != nil {
		if err == parse.ErrNotParseable {
			fmt.Fprintf(os.Stderr, "tparse error: no parseable events: call go test with -json flag\n\n")
			if *dumpPtr {
				parse.ReplayOutput(os.Stderr, &replayBuf)
			}
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "tparse error: %v\n\n", err)
		parse.ReplayOutput(os.Stderr, &replayBuf)
		os.Exit(1)
	}

	if len(pkgs) == 0 {
		fmt.Fprintf(os.Stdout, "tparse: no go packages to parse\n\n")
		parse.ReplayOutput(os.Stderr, &replayBuf)
		os.Exit(1)
	}

	// Use this value to print to stdout (0) or stderr (>=1)
	exitCode := pkgs.ExitCode()

	w := newWriter(exitCode)

	opts := testsTableOptions{
		trim: *smallScreenPtr,
	}
	if *allPtr {
		opts.pass, opts.skip = true, true
	} else if *passPtr {
		opts.pass, opts.skip = true, false
	} else if *skipPtr {
		opts.pass, opts.skip = false, true
	}

	if *topPtr {
		w.SummaryTable(pkgs, *showNoTestsPtr)
		w.PrintFailed(pkgs)
		w.TestsTable(pkgs, opts)
		if *dumpPtr {
			parse.ReplayOutput(os.Stderr, &replayBuf)
		}
	} else {
		// Default.
		if *dumpPtr {
			parse.ReplayOutput(os.Stderr, &replayBuf)
		}
		w.TestsTable(pkgs, opts)
		w.PrintFailed(pkgs)
		w.SummaryTable(pkgs, *showNoTestsPtr)
	}

	// Return proper exit code. This must be consistent with what go test would have
	// returned without tparse.
	os.Exit(exitCode)
}

// newWriter initializes a console writer based on a given exit code.
// 0 writes to stdout, >=1 writes to stderr
func newWriter(exitCode int) *consoleWriter {
	w := consoleWriter{
		Color:  !*noColorPtr, // Color enabled by default.
		Output: colorable.NewColorableStdout(),
	}

	// retrn all output for non-zero exit codes to std err
	if exitCode != 0 {
		w.Output = colorable.NewColorableStderr()
	}

	return &w
}

// newReader returns a reader; either a named pipe or open file.
func newReader() (io.ReadCloser, error) {

	switch flag.NArg() {
	case 0: // Get FileInfo interface and fail everything except a named pipe (FIFO).

		finfo, err := os.Stdin.Stat()

		if err != nil {
			return nil, err
		}

		// Check file mode bits to test for named pipe as stdin.
		if finfo.Mode()&os.ModeNamedPipe != 0 {
			return os.Stdin, nil
		}

		return nil, errors.New("when no files are supplied as arguments stdin must be a named pipe")

	default: // Attempt to read from a file.
		f, err := os.Open(os.Args[len(os.Args)-flag.NArg()]) // 🦄
		if err != nil {
			return nil, err
		}

		return f, nil
	}
}

func (w *consoleWriter) SummaryTable(pkgs parse.Packages, showNoTests bool) {
	fmt.Fprintln(w.Output)

	tbl := tablewriter.NewWriter(w.Output)
	tbl.SetHeader([]string{
		"Status",  // 0
		"Elapsed", // 1
		"Package", // 2
		"Cover",   // 3
		"Pass",    // 4
		"Fail",    // 5
		"Skip",    // 6
	})

	tbl.SetAutoWrapText(false)

	var passed [][]string
	var notests [][]string

	for name, pkg := range pkgs {

		var elapsed string
		if pkg.Cached {
			elapsed = "(cached)"
		} else {
			elapsed = strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s"
		}

		if pkg.HasPanic {
			tbl.Append([]string{
				colorize("PANIC", cRed, w.Color), elapsed, name, "--", "--", "--", "--",
			})
			continue
		}

		if pkg.NoTestFiles {
			notests = append(notests, []string{
				colorize("NOTEST", cYellow, w.Color), elapsed, name + "\n[no test files]", "--", "--", "--", "--",
			})
			continue
		}

		if pkg.NoTests {
			if len(pkg.NoTestSlice) > 0 {
				// This should capture cases where packages have a mixture of empty and non-empty test files.
				var ss []string
				for i, t := range pkg.NoTestSlice {
					i++
					ss = append(ss, fmt.Sprintf("%d.%s", i, t.Test))
				}
				s := fmt.Sprintf("%s\n[no tests to run]\n%s", name, strings.Join(ss, "\n"))
				notests = append(notests, []string{
					colorize("NOTEST", cYellow, w.Color), elapsed, s, "--", "--", "--", "--",
				})

				if len(pkg.TestsByAction(parse.ActionPass)) == len(pkg.NoTestSlice) {
					continue
				}

			} else {
				// This should capture cases where packages truly have no tests, but empty files.
				notests = append(notests, []string{
					colorize("NOTEST", cYellow, w.Color), elapsed, name + "\n[no tests to run]", "--", "--", "--", "--",
				})
				continue
			}
		}

		coverage := fmt.Sprintf("%.1f%%", pkg.Coverage)
		if pkg.Summary.Action != parse.ActionFail {
			switch c := pkg.Coverage; {
			case c == 0.0:
				break
			case c <= 50.0:
				coverage = colorize(coverage, cRed, w.Color)
			case pkg.Coverage > 50.0 && pkg.Coverage < 80.0:
				coverage = colorize(coverage, cYellow, w.Color)
			case pkg.Coverage >= 80.0:
				coverage = colorize(coverage, cGreen, w.Color)
			}
		}

		passed = append(passed, []string{
			withColor(pkg.Summary.Action, w.Color), //0
			elapsed,                                //1
			name,                                   //2
			coverage,                               //3
			strconv.Itoa(len(pkg.TestsByAction(parse.ActionPass))), //4
			strconv.Itoa(len(pkg.TestsByAction(parse.ActionFail))), //5
			strconv.Itoa(len(pkg.TestsByAction(parse.ActionSkip))), //6
		})
	}

	if tbl.NumLines() == 0 && len(passed) == 0 && len(notests) == 0 {
		return
	}

	if len(passed) > 0 {
		tbl.AppendBulk(passed)
		if showNoTests {
			// Only display the "no tests to run" cases if users want to see them when passed
			// tests are available.
			tbl.AppendBulk(notests)
		}
	} else {
		tbl.AppendBulk(notests)
	}

	tbl.Render()
}

type testsTableOptions struct {
	pass, skip, trim bool
}

func (w *consoleWriter) TestsTable(pkgs parse.Packages, options testsTableOptions) {
	// Print passed tests, sorted by elapsed. Unlike failed tests, passed tests
	// are not grouped. Maybe bad design?
	tbl := tablewriter.NewWriter(w.Output)

	tbl.SetHeader([]string{
		"Status",
		"Elapsed",
		"Test",
		"Package",
	})

	tbl.SetAutoWrapText(false)

	var sp []*parse.Package

	for _, pkg := range pkgs {
		if pkg.NoTestFiles || pkg.NoTests || pkg.HasPanic {
			continue
		}
		sp = append(sp, pkg)
	}

	numPkgs := len(sp)
	numScanned := 0

	for _, pkg := range sp {
		numScanned++

		var all []*parse.Test
		if options.skip {
			skipped := pkg.TestsByAction(parse.ActionSkip)
			all = append(all, skipped...)
		}
		if options.pass {
			passed := pkg.TestsByAction(parse.ActionPass)

			// Sort tests within a package by elapsed time in descending order, longest on top.
			sort.Slice(passed, func(i, j int) bool {
				return passed[i].Elapsed() > passed[j].Elapsed()
			})

			all = append(all, passed...)
		}
		if len(all) == 0 {
			continue
		}

		for _, t := range all {
			t.SortEvents()

			var testName strings.Builder
			testName.WriteString(t.Name)
			if options.trim && testName.Len() > 32 && strings.Count(testName.String(), "/") > 0 {
				testName.Reset()
				ss := strings.Split(t.Name, "/")
				testName.WriteString(ss[0] + "\n")
				for i, s := range ss[1:] {
					testName.WriteString(" /" + s)
					if i != len(ss[1:])-1 {
						testName.WriteString("\n")
					}
				}
			}

			tbl.Append([]string{
				withColor(t.Status(), w.Color),
				strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName.String(),
				filepath.Base(t.Package),
			})
		}

		// Add empty line between package groups except the last package
		if numScanned < numPkgs {
			tbl.Append([]string{"", "", "", ""})
		}
	}

	if tbl.NumLines() > 0 {
		fmt.Fprintf(w.Output, "\n")
		tbl.Render()
	}
}

func (w *consoleWriter) PrintFailed(pkgs parse.Packages) {
	// Print all failed tests per package (if any). Panic is an exception.
	for _, pkg := range pkgs {

		if pkg.HasPanic {
			// may or may not be associated with tests, so we print it separately.
			w.PrintPanic(pkg)
			continue
		}

		failed := pkg.TestsByAction(parse.ActionFail)
		if len(failed) == 0 {
			continue
		}

		s := fmt.Sprintf("\nFAIL: %s", pkg.Summary.Package)
		n := make([]string, len(s))
		sn := fmt.Sprintf("%s\n%s\n", s, strings.Join(n, "-"))

		fmt.Fprintf(w.Output, colorize(sn, cRed, w.Color))

		for i, t := range failed {
			t.SortEvents()

			fmt.Fprintf(w.Output, "%s", t.Stack())
			if i < len(failed)-1 {
				fmt.Fprintf(w.Output, "\n")
			}
		}
	}
}

func (w *consoleWriter) PrintPanic(pkg *parse.Package) {
	s := fmt.Sprintf("\nPANIC: %s: %s", pkg.Summary.Package, pkg.Summary.Test)
	n := make([]string, len(s)+1)
	sn := fmt.Sprintf("%s\n%s\n", s, strings.Join(n, "-"))
	fmt.Fprintf(w.Output, colorize(sn, cRed, w.Color))

	for _, e := range pkg.PanicEvents {
		fmt.Fprint(w.Output, e.Output)
	}
}

// withColor attempts to return a colorized string based on action if enabled:
// pass=green, skip=yellow, fail=red, default=no color.
func withColor(a parse.Action, enabled bool) string {
	s := strings.ToUpper(a.String())
	if !enabled {
		return s
	}
	switch a {
	case parse.ActionPass:
		return colorize(s, cGreen, true)
	case parse.ActionSkip:
		return colorize(s, cYellow, true)
	case parse.ActionFail:
		return colorize(s, cRed, true)
	default:
		return s
	}
}

const (
	cReset  = 0
	cRed    = 31
	cGreen  = 32
	cYellow = 33
)

func colorize(s string, color int, enabled bool) string {
	if !enabled {
		return s
	}
	return fmt.Sprintf("\x1b[1;%dm%s\x1b[0m", color, s)
}
