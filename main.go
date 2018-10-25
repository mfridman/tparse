package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/mfridman/tparse/parse"
	"github.com/olekukonko/tablewriter"

	"github.com/pkg/errors"
)

// Flags.
var (
	versionPtr = flag.Bool("v", false, "")
	allPtr     = flag.Bool("all", false, "")
	passPtr    = flag.Bool("pass", false, "")
	skipPtr    = flag.Bool("skip", false, "")
	noTestsPtr = flag.Bool("notests", false, "")
	dumpPtr    = flag.Bool("dump", false, "")
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
	-skip		Display table for passed skipped tests.
	-notests	Display packages containing no test files in summary.
	-dump		Enables recovering initial go test output in non-JSON format following Summary and Test tables.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprint(usage))
		os.Exit(2)
	}
	flag.Parse()

	if *versionPtr {
		fmt.Println("tparse version: devel")
		os.Exit(0)
	}

	r, err := getReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flag.Usage()
	}
	defer r.Close()

	pkgs, err := parse.Start(r)
	// TODO(mf): no matter what error we get, we should always allow the user to retrieve
	// whatever we could read in Start with -dump. Currently it only gets called way below.
	if err != nil {
		switch err := errors.Cause(err).(type) {
		case *json.SyntaxError:
			fmt.Fprint(os.Stderr, "Error: must call go test with -json flag\n\n")
			flag.Usage()
		case *parse.PanicErr:
			// Just return the package name, test name and debug info from the panic.
			err.PrintPanic()
			os.Exit(1)
		default:
			// TODO(mf):
			// - Does it make sense to display error and usage
			// back to the user when there is a scan error?
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			flag.Usage()
		}
	}

	if len(pkgs) == 0 {
		parse.RawDump()
		os.Exit(0)
	}

	// Prints packages summary table.
	// TODO: think about using functional options?
	pkgs.PrintSummary(*noTestsPtr)

	// Print all failed tests per package (if any).
	for _, p := range pkgs {
		failed := p.TestsByAction(parse.ActionFail)
		if len(failed) == 0 {
			continue
		}

		s := fmt.Sprintf("PACKAGE: %s", p.Summary.Package)
		n := make([]string, len(s)+1)
		fmt.Printf("%s\n%s\n", s, strings.Join(n, "-"))

		for _, t := range failed {
			t.PrintFail()
		}
	}

	if *allPtr || *passPtr {
		// Print passed tests, sorted by elapsed. Unlike failed tests, passed tests
		// are not grouped. Maybe bad design?
		tbl := tablewriter.NewWriter(os.Stdout)

		tbl.SetHeader([]string{
			"Status",
			"Elapsed",
			"Test Name",
			"Package",
		})

		tbl.SetColMinWidth(3, 40)

		var i int
		for _, p := range pkgs {
			passed := p.TestsByAction(parse.ActionPass)
			if len(passed) == 0 {
				continue
			}

			// Sort tests within a package by elapsed time in descending order, longest on top.
			sort.Slice(passed, func(i, j int) bool {
				return passed[i].Elapsed() > passed[j].Elapsed()
			})

			for _, t := range passed {
				tbl.Append(t.PrintPass())
			}

			// Add empty lines between tests of packages. Maybe tablewriter has some feature, but this is easier (for now).
			if i != len(pkgs)-1 {
				tbl.Append([]string{"", "", "", ""})
			}
			i++

		}

		tbl.Render()
	}

	if *dumpPtr {
		parse.RawDump()
	}

	// Return an exit code that's inline with what go test would have returned otherwise.
	for _, p := range pkgs {
		if p.Summary.Action == parse.ActionFail {
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
