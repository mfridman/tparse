package parse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// ErrNotParseable indicates the event line was not parseable. It is returned only
// by the Process func.
var ErrNotParseable = errors.New("failed to parse events")

// ErrRaceDetected indicates a race condition has been detected during execution.
// Returned by the Process func.
var ErrRaceDetected = errors.New("race detected")

// Process is the entry point to the parse pkg. It consumes a reader
// and attempts to parse go test JSON output lines until EOF.
//
// Note, Process will attempt to parse up to 50 lines before returning an error.
//
// Returns PanicErr on the first package containing a test that panics.
func Process(r io.Reader) (Packages, error) {

	pkgs := Packages{}

	var hasRace bool

	var scan bool
	var badLines int

	sc := bufio.NewScanner(r)
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
		// Short circuit output when panic is detected.
		if pkg.HasPanic {
			pkg.PanicEvents = append(pkg.PanicEvents, e)
			continue
		}

		if e.IsRace() {
			hasRace = true
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
	if hasRace {
		return nil, ErrRaceDetected
	}

	return pkgs, nil
}

// ReplayOutput takes json event lines from r and returns output actions to w.
// If an error occurs parsing an event and the output action cannot be retrieved
// the raw line of text is printed.
//
// Used to parse JSON lines into their raw output, i.e., what go test output
// would have been without -json.
func ReplayOutput(w io.Writer, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			// We couldn't parse an event, so return the raw text.
			fmt.Fprintln(w, strings.TrimSpace(sc.Text()))
			continue
		}
		// Output expected to be terminated by a newline.
		fmt.Fprint(w, e.Output)
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(w, "tparse scan error: %v\n", err)
	}
}

// ReplayRaceOutput takes json event lines from r and returns partial output
// to w. Specifically, once a race is detected all discard and PASS events will
// be ignored. This is to keep output as close as possible to what
// go test (without -v) would have otherwise returned.
//
// The race output is non-detertministc.
// https://github.com/golang/go/issues/29156#issuecomment-445486381
func ReplayRaceOutput(w io.Writer, r io.Reader) {

	var raceStarted bool

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			// We couldn't parse an event, so return the raw text.
			fmt.Fprintln(w, strings.TrimSpace(sc.Text()))
			continue
		}

		if raceStarted {
			if e.Discard() || strings.Contains(e.Output, "--- PASS:") {
				continue
			}
			fmt.Fprint(w, e.Output)
			continue
		}

		if strings.Contains(e.Output, "==================") {
			raceStarted = true
			fmt.Fprint(w, e.Output)
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(w, "tparse scan error: %v\n", err)
	}
}
