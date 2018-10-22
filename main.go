package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
)

var usage = `Usage:
	go test ./... -json | tparse [options...]
	go test [packages...] -json | tparse [options...]
	go test [packages...] -json > pkgs.out ; tparse [options...] pkgs.out

Options:
	-h		Show help.
	-v		Show version.
	-all		Display all event types: pass, skip and fail. (Failed items are always displayed)
	-pass		Display all passed tests.
	-skip		Display all skipped tests.
	-notests	Display packages with no tests in summary.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprint(usage))
		os.Exit(2)
	}
	flag.Parse()

	if *versionPtr {
		fmt.Println("tparse version 0.0.1")
		os.Exit(0)
	}

	log.SetFlags(0)

	r, err := getReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flag.Usage()
	}

	pkgs, err := parse.Do(r)
	if err != nil {
		switch err := errors.Cause(err).(type) {
		case *json.SyntaxError:
			fmt.Fprint(os.Stderr, "Error: must call go test with -json flag\n\n")
			flag.Usage()
		case *parse.PanicErr:
			err.PrintPanic()
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			flag.Usage()
		}
	}

	// Prints packages summary table.
	// TODO: think about using functional options?
	pkgs.Print(*noTestsPtr)

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

	// Return an exit code that's inline with what go test would have returned otherwise.
	// TODO: validate this is true, if at least one package is failed the exit code is set to 1.
	// for _, p := range pkgs {
	// 	if p.Summary.Action == parse.ActionFail {
	// 		os.Exit(1)
	// 	}
	// }
}

// read from a named pipe (no args) or from single arg expected to be a faile path
func getReader() (io.Reader, error) {

	switch flag.NArg() {
	case 0: // Get FileInfo interface and fail everything except a named pipe (FIFO).

		finfo, err := os.Stdin.Stat()

		if err != nil {
			return nil, err
		}

		// check file mode bits to test for named pipe as stdin
		if finfo.Mode()&os.ModeNamedPipe != 0 {
			dat, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(dat), nil
		}

		return nil, errors.New("when no files are supplied as arguments stdin must be a named pipe")

	default: // Attempt to read from a file.

		dat, err := ioutil.ReadFile(os.Args[len(os.Args)-flag.NArg()]) // 🦄
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(dat), nil
	}
}
