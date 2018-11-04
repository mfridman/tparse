package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mfridman/tparse/version"
	"github.com/olekukonko/tablewriter"

	"github.com/mfridman/tparse/parse"

	"github.com/pkg/errors"
)

// Flags.
var (
	versionPtr     = flag.Bool("v", false, "")
	allPtr         = flag.Bool("all", false, "")
	passPtr        = flag.Bool("pass", false, "")
	skipPtr        = flag.Bool("skip", false, "")
	showNoTestsPtr = flag.Bool("notests", false, "")
	dumpPtr        = flag.Bool("dump", false, "")
	smallScreenPtr = flag.Bool("smallscreen", false, "")
	topPtr         = flag.Bool("top", false, "")
)

var usage = `Usage:
	go test ./... -json | tparse [options...]
	go test [packages...] -json | tparse [options...]
	go test [packages...] -json > pkgs.out ; tparse [options...] pkgs.out

Options:
	-h		Show help.
	-v		Show version.
	-all		Display table event for pass, skip and fail. (Failed items are always displayed)
	-pass		Display table for passed tests.
	-skip		Display table for skipped tests.
	-notests	Display packages containing no test files or empty test files in summary.
	-dump		Enables recovering go test output in non-JSON format.
	-smallscreen	Split subtest names vertically to fit on smaller screens.
	-top		Display summary table towards top.
`

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

	r, err := getReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flag.Usage()
	}
	defer r.Close()

	pkgs, err := parse.Process(r)
	// TODO(mf): no matter what error we get, we should always allow the user to retrieve
	// whatever Process was able to read with -dump. Currently it gets called way below.
	if err != nil {
		if err == parse.ErrNotParseable {
			fmt.Fprintf(os.Stderr, "tparse error: no parseable events: call go test with -json flag\n")
			os.Exit(1)
		}

		// TODO(mf):
		// - Does it make sense to display error and usage
		// back to the user when there is a scan error?
		fmt.Fprintf(os.Stderr, "tparse error: %v\n\n", err)
		parse.RawDump(os.Stderr, *dumpPtr)
		flag.Usage()
	}

	if len(pkgs) == 0 {
		parse.RawDump(os.Stderr, true)
		os.Exit(0)
	}

	if *topPtr {
		printSummary(os.Stdout, pkgs, *showNoTestsPtr)
	}

	parse.RawDump(os.Stderr, *dumpPtr)

	// Print all failed tests per package (if any).
	printFailed(os.Stderr, pkgs)

	if *allPtr {
		printTests(os.Stdout, pkgs, true, true, *smallScreenPtr)
	} else if *passPtr {
		printTests(os.Stdout, pkgs, true, false, *smallScreenPtr)
	} else if *skipPtr {
		printTests(os.Stdout, pkgs, false, true, *smallScreenPtr)
	}

	// Prints packages summary table.
	// TODO: think about using functional options?
	if !*topPtr {
		printSummary(os.Stdout, pkgs, *showNoTestsPtr)
	}

	// Return an exit code that's inline with what go test would have returned otherwise.
	for _, p := range pkgs {
		if p.HasPanic || p.Summary.Action == parse.ActionFail {
			os.Exit(1)
		}
	}
}

// getReader returns a reader; either a named pipe or open file.
func getReader() (io.ReadCloser, error) {

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

func printSummary(w io.Writer, pkgs parse.Packages, showNoTests bool) {
	fmt.Fprintf(w, "\n")

	tbl := tablewriter.NewWriter(w)
	tbl.SetHeader([]string{
		"Status",  //0
		"Elapsed", //1
		"Package", //2
		"Cover",   //3
		"Pass",    //4
		"Fail",    //5
		"Skip",    //6
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
				colorize("PANIC", cRed, true), elapsed, name, "--", "--", "--", "--",
			})
			continue
		}

		if pkg.NoTestFiles {
			notests = append(notests, []string{
				colorize("NOTEST", cYellow, true), elapsed, name + "\n[no test files]", "--", "--", "--", "--",
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
					colorize("NOTEST", cYellow, true), elapsed, s, "--", "--", "--", "--",
				})

				if len(pkg.TestsByAction(parse.ActionPass)) == len(pkg.NoTestSlice) {
					continue
				}

			} else {
				// This should capture cases where packages truly have no tests, but empty files.
				notests = append(notests, []string{
					colorize("NOTEST", cYellow, true), elapsed, name + "\n[no tests to run]", "--", "--", "--", "--",
				})
				continue
			}
		}

		coverage := fmt.Sprintf("%.1f%%", pkg.Coverage)
		switch c := pkg.Coverage; {
		case c == 0.0:
			break
		case c <= 50.0:
			coverage = colorize(coverage, cRed, true)
		case pkg.Coverage > 50.0 && pkg.Coverage < 80.0:
			coverage = colorize(coverage, cYellow, true)
		case pkg.Coverage >= 80.0:
			coverage = colorize(coverage, cGreen, true)
		}

		passed = append(passed, []string{
			withColor(pkg.Summary.Action), //0
			elapsed,                       //1
			name,                          //2
			coverage,                      //3
			strconv.Itoa(len(pkg.TestsByAction(parse.ActionPass))), //4
			strconv.Itoa(len(pkg.TestsByAction(parse.ActionFail))), //5
			strconv.Itoa(len(pkg.TestsByAction(parse.ActionSkip))), //6
		})
	}

	if len(passed) == 0 && len(notests) == 0 {
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

func printFailed(w io.Writer, pkgs parse.Packages) {
	// Print all failed tests per package (if any). Panic is an exception.
	for _, pkg := range pkgs {

		if pkg.HasPanic {
			// may or may not be associated with tests, so we print it separately.
			printPanic(pkg, os.Stderr)
			continue
		}

		failed := pkg.TestsByAction(parse.ActionFail)
		if len(failed) == 0 {
			continue
		}

		s := fmt.Sprintf("\nFAIL: %s", pkg.Summary.Package)
		n := make([]string, len(s))
		sn := fmt.Sprintf("%s\n%s\n", s, strings.Join(n, "-"))

		fmt.Fprintf(w, colorize(sn, cRed, true))

		for i, t := range failed {
			t.SortEvents()

			fmt.Fprintf(w, "%s", t.Stack())
			if i < len(failed)-1 {
				fmt.Fprintf(w, "\n")
			}
		}
	}
}

func printPanic(pkg *parse.Package, w io.Writer) {
	s := fmt.Sprintf("\nPANIC: %s: %s", pkg.Summary.Package, pkg.Summary.Test)
	n := make([]string, len(s)+1)
	sn := fmt.Sprintf("%s\n%s\n", s, strings.Join(n, "-"))
	fmt.Fprintf(w, colorize(sn, cRed, true))

	for _, e := range pkg.PanicEvents {
		fmt.Fprint(w, e.Output)
	}
}

func printTests(w io.Writer, pkgs parse.Packages, pass, skip, trim bool) {
	fmt.Fprintf(w, "\n")

	// Print passed tests, sorted by elapsed. Unlike failed tests, passed tests
	// are not grouped. Maybe bad design?
	tbl := tablewriter.NewWriter(w)

	tbl.SetHeader([]string{
		"Status",
		"Elapsed",
		"Test",
		"Package",
	})

	tbl.SetAutoWrapText(false)

	for _, pkg := range pkgs {
		if pkg.NoTestFiles {
			continue
		}

		var all []*parse.Test
		if skip {
			skipped := pkg.TestsByAction(parse.ActionSkip)
			all = append(all, skipped...)
		}
		if pass {
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
			if trim && testName.Len() > 32 && strings.Count(testName.String(), "/") > 0 {
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
				withColor(t.Status()),
				strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName.String(),
				filepath.Base(t.Package),
			})
		}

		// Add empty line between package groups.
		// TODO(mf): don't add line to last item
		tbl.Append([]string{"", "", "", ""})
	}

	if tbl.NumLines() > 0 {
		tbl.Render()
	}
}

// withColor attempts to return a colorized string based on action:
// pass=green, skip=yellow, fail=red, default=no color.
func withColor(a parse.Action) string {
	s := strings.ToUpper(a.String())
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
