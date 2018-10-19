package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/mfridman/tparse/parse"
	"github.com/olekukonko/tablewriter"

	"github.com/pkg/errors"
)

// TODO: if a user reruns an unmodified test or package, it's considered cached. This should be displayed in some way.
// -count=1 to force a run (no cache)

// Flags.
var (
	versionPtr = flag.Bool("v", false, "")
	allPtr     = flag.Bool("all", false, "")
	passPtr    = flag.Bool("pass", false, "")
	skipPtr    = flag.Bool("skip", false, "")
)

var usage = `Usage:
	go test [packages...] -json | tparse [options...]
	go test [packages...] -json > pkgs.out ; tparse [options...] pkgs.out

Options:
	-h	Show help.
	-v	Show version.
	-all	Display all event types: pass, skip and fail. (Failed items are always displayed)
	-pass	Display all passed tests.
	-skip	Display all skipped tests.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprint(usage))
	}
	flag.Parse()

	if *versionPtr {
		fmt.Println("tparse version 0.0.1")
		os.Exit(0)
	}

	log.SetFlags(0)
	log.SetPrefix("tparse error: ")

	r, err := getReader()
	if err != nil {
		log.Fatal(err)
	}

	pkgs, err := parse.Do(r)
	if err != nil {
		// TODO: if anything goes wrong parsing, we need to return back whatever user has piped in.
		// assuming we were able to get a reader from getReader
		// Also need to handle panics.
		log.Fatal(err)
	}

	// Prints the top-most summary box.
	pkgs.Print()

	// Print all failed tests per package (if any).
	for _, p := range pkgs {
		failed := p.TestsByAction(parse.ActionFail)
		if len(failed) == 0 {
			continue
		}

		fmt.Printf("%s\n", p.Summary.Package)

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

	// TODO: we need to return an exit code that's inline with what go test would have returned.
}

// read from a named pipe (no args) or from single arg expected to be a faile path
func getReader() (io.Reader, error) {

	switch len(os.Args[1:]) {
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

		return nil, errors.Errorf("stdin must be a named pipe (FIFO). current filemode: %q\n", finfo.Mode())

	default: // Attempt to read from a file.

		// After processing all flags, we should have one file to read from, fail otherwise.
		if flag.NArg() < 1 {
			flag.Usage()
			os.Exit(1)
		}

		dat, err := ioutil.ReadFile(os.Args[len(os.Args)-flag.NArg()]) // 🦄
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(dat), nil
	}
}
