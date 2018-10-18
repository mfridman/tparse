package main

import (
	"bytes"
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

func main() {

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

	// Print failed package tests (if any).
	for _, p := range pkgs {
		failed := p.Failed()
		if len(failed) == 0 {
			continue
		}

		fmt.Printf("%s\n", p.Summary.Package)

		for _, t := range failed {
			t.PrintFail()
		}
	}

	// Print passed tests, sorted by elapsed. Unlike failed tests, passed tests
	// are not grouped. Maybe bad design?
	tbl := tablewriter.NewWriter(os.Stdout)

	tbl.SetHeader([]string{
		"Status",
		"Elapsed",
		"Test Name",
		"Package",
	})

	var i int
	for _, p := range pkgs {
		passed := p.Passed()
		if len(passed) == 0 {
			continue
		}

		// Sort tests within a package by elapsed time in descending order, longest on top.
		// TODO: I don't like how this works. Ideally all "pass" tests of all packages will be grouped
		// and sorted by elapsed time.
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

	// TODO: we need to return an exit code that's inline with what go test would have returned.
}

// read from a named pipe (no args) or from single arg expected to be a faile path
func getReader() (io.Reader, error) {

	switch len(os.Args[1:]) {
	case 0: // get FileInfo interface and fail everything except a named pipe (FIFO)

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

	case 1: // read from first arg, which is a file

		dat, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(dat), nil

	default:

		return nil, errors.New("input must be a single file or stdin")
	}
}
