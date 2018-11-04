package parse

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// ErrNotParseable indicates the event line was not parseable. It is returned only
// by the Process func.
var ErrNotParseable = errors.New("failed to parse events")

// Packages is a collection of packages being tested.
// TODO: this should really be a consoleWriter... would benefit from a nice refactor.
type Packages map[string]*Package

// Process is the entry point to the parse pkg. It consumes a reader
// and attempts to parse go test JSON output lines until EOF.
//
// Note, Process will attempt to parse up to 50 lines before returning an error.
//
// Returns PanicErr on the first package containing a test that panics.
func Process(r io.Reader) (Packages, error) {

	pkgs := Packages{}

	var scan bool
	var badLines int

	tr := io.TeeReader(r, &rawdump)

	sc := bufio.NewScanner(tr)
	for sc.Scan() {
		// Scan up-to 50 lines for a parseable event, if we get one, expect
		// no errors to follow until EOF.
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			badLines++
			if scan || badLines > 50 {
				switch err.(type) {
				case *json.SyntaxError:
					return nil, ErrNotParseable
				default:
					return nil, err
				}
			}
			continue
		}
		scan = true

		pkg, ok := pkgs[e.Package]
		if !ok {
			pkg = NewPackage()
			pkgs[e.Package] = pkg
		}

		if e.IsPanic() {
			pkg.HasPanic = true
			pkg.Summary.Action = ActionFail
			pkg.Summary.Package = e.Package
			pkg.Summary.Test = e.Test
		}
		if pkg.HasPanic {
			pkg.PanicEvents = append(pkg.PanicEvents, e)
			continue
		}

		if e.IsCached() {
			pkg.Cached = true
		}

		if e.NoTestFiles() {
			pkg.NoTestFiles = true
			// Manually mark [no test files] as "pass", because the go test tool reports the
			// package Summary action as "skip".
			pkg.Summary.Package = e.Package
			pkg.Summary.Action = ActionPass
		}
		if e.NoTestsWarn() {
			// One or more tests within the package contains no tests.
			pkg.NoTestSlice = append(pkg.NoTestSlice, e)
		}

		if e.NoTestsToRun() {
			// Only packages marked as "pass" will contain a summary line appended with [no tests to run].
			// This indicates one or more tests is marked as having no tests to run.
			pkg.NoTests = true
			pkg.Summary.Package = e.Package
			pkg.Summary.Action = ActionPass
		}

		if e.LastLine() {
			pkg.Summary = e
			continue
		}

		cover, ok := e.Cover()
		if ok {
			pkg.Cover = true
			pkg.Coverage = cover
		}

		if !e.Discard() {
			pkg.AddEvent(e)
		}
	}

	if err := sc.Err(); err != nil {
		return nil, errors.Wrap(err, "bufio scanner error")
	}
	if !scan {
		return nil, ErrNotParseable
	}

	return pkgs, nil
}

// rawdump is written to within the Process function. It holds all original incoming
// events.
var rawdump bytes.Buffer

// RawDump prints back all lines that Process func reads into the specified writer.
// Each line is parsed as an Event and output is printed. If an error occurs
// parsing an event the raw line of text is printed.
func RawDump(w io.Writer, dump bool) {
	if !dump {
		return
	}
	fmt.Fprintf(w, "\n")

	sc := bufio.NewScanner(&rawdump)
	for sc.Scan() {
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			// We couldn't parse an event, so return the raw text.
			fmt.Fprintln(w, strings.TrimSpace(sc.Text()))
			continue
		}
		fmt.Fprint(w, e.Output)
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(w, "tparse scan error: %v\n", err)
	}
}
