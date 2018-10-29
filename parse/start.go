package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Start is the entry point to the parse pkg. It consumes a reader
// and attempts to parse go test JSON output lines until EOF.
//
// Note, Start will attempt to parse up to 50 lines before failing.
//
// Returns PanicErr on the first package containing a test that panics.
func Start(r io.Reader) (Packages, error) {

	pkgs := Packages{}

	var scan bool
	var badLines int

	rd := io.TeeReader(r, &rawdump)

	sc := bufio.NewScanner(rd)
	for sc.Scan() {

		// We'll "prescan" up-to 50 lines for a parsable event. If we do get a parsable event
		// we expect no errors to follow until EOF. For each unparsable event we will print
		// up to 50 lines to stderr.
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			if scan || badLines >= 50 {
				return nil, err
			}

			// TODO(mf): do we want to return unparsable events?
			// This should be available downstream through RawDump.
			// Alternatively, instead of RawDump we could just print those lines
			// out start recording once we have a good event?

			// fmt.Fprintln(os.Stderr, sc.Text())
			badLines++
			continue
		}
		scan = true

		pkg, ok := pkgs[e.Package]
		if !ok {
			pkg = NewPackage()
			pkgs[e.Package] = pkg
		}

		if e.NoTestFiles() {
			pkg.Summary = &Event{Action: ActionPass}
			pkg.NoTestFiles = true
		}

		if e.NoTestsWarn() {
			pkg.NoTestSlice = append(pkg.NoTestSlice, e)
			// The package summary line within NoTestsToRun will mark the package as [no tests to run].
		}
		if e.NoTestsToRun() {
			pkg.Summary = &Event{Action: ActionPass}
			pkg.NoTests = true
		}

		if e.IsCached() {
			pkg.Cached = true
		}

		cover, ok := e.Cover()
		if ok {
			pkg.Cover = true
			pkg.Coverage = cover
		}

		if e.Summary() {
			pkg.Summary = e
			continue
		}

		// We don't need to save these line.
		if e.Discard() {
			continue
		}

		pkg.AddEvent(e)
	}

	if err := sc.Err(); err != nil {
		return nil, errors.Wrap(err, "bufio scanner error")
	}

	// Panic means end of the world, return PanicErr.
	for _, pkg := range pkgs {
		if err := pkg.HasPanic(); err != nil {
			return nil, err
		}
	}

	return pkgs, nil
}

// rawdump is written to within the Start function. It holds all original incoming
// events.
var rawdump bytes.Buffer

// RawDump prints back all lines that Start func reads. Each line is parsed as an Event
// and output is printed. If an error occurs parsing an event the raw line of text is printed.
func RawDump() {
	sc := bufio.NewScanner(&rawdump)
	for sc.Scan() {
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			// We couldn't parse an event, so return the raw text.
			fmt.Fprintln(os.Stderr, strings.TrimSpace(sc.Text()))
			continue
		}
		fmt.Fprint(os.Stderr, e.Output)
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "tparse scan error: %v\n", err)
	}
}
