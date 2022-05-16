package parse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

// ErrNotParseable indicates the event line was not parseable.
var ErrNotParseable = errors.New("failed to parse")

// ErrRaceDetected indicates a race condition has been detected during execution.
// Returned by the Process func.
var ErrRaceDetected = errors.New("race detected")

type options struct {
	w      io.Writer
	follow bool
	debug  bool
}

type OptionsFunc func(o *options)

func WithFollowOutput(b bool) OptionsFunc {
	return func(o *options) { o.follow = b }
}

func WithWriter(w io.Writer) OptionsFunc {
	return func(o *options) { o.w = w }
}

func WithDebug() OptionsFunc {
	return func(o *options) { o.debug = true }
}

// Process is the entry point to the parse pkg. It consumes a reader
// and attempts to parse go test JSON output lines until EOF.
//
// Note, Process will attempt to parse up to 50 lines before returning an error.
//
// Returns PanicErr on the first package containing a test that panics.
func Process(r io.Reader, opts ...OptionsFunc) (Packages, error) {
	option := &options{}
	for _, f := range opts {
		f(option)
	}

	provider := &provider{
		packages: make(Packages),
	}

	sc := bufio.NewScanner(r)
	var started bool
	var badLines int
	for sc.Scan() {
		// Scan up-to 50 lines for a parseable event, if we get one, expect
		// no errors to follow until EOF.
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			badLines++
			if started || badLines > 50 {
				switch err.(type) {
				case *json.SyntaxError:
					return nil, fmt.Errorf("line %d json error: %s: %w", badLines, err.Error(), ErrNotParseable)
				default:
					return nil, err
				}
			}
			if option.follow && option.w != nil {
				fmt.Fprintf(option.w, "%s\n", sc.Bytes())
			}
			continue
		}
		started = true
		// Optionally, as test output is piped to us we write the plain
		// text Output as if go test was run without the -json flag.
		if option.follow && option.w != nil {
			fmt.Fprint(option.w, e.Output)
		}

		provider.addEventToPackage(e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("received scanning error: %w", err)
	}
	// Entire input has been scanned and no go test json output was found.
	if !started {
		return nil, ErrNotParseable
	}

	return provider.packages, nil
}

type provider struct {
	packages     Packages
	raceDetected bool
}

func (p *provider) addEventToPackage(e *Event) {
	pkg, ok := p.packages[e.Package]
	if !ok {
		pkg = newPackage()
		p.packages[e.Package] = pkg
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
		return
	}
	if e.IsRace() {
		p.raceDetected = true
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
		return
	}
	if cover, ok := e.Cover(); ok {
		pkg.Cover = true
		pkg.Coverage = cover
	}
	if !e.Discard() {
		pkg.AddEvent(e)
	}
}

// ReplayRaceOutput takes json event lines from r and returns partial output
// to w. Specifically, once a race is detected all PASS events and update events
// will be ignored. This is to keep output as close as possible to what
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
			for i := range updates {
				if strings.HasPrefix(e.Output, updates[i]) || strings.Contains(e.Output, "--- PASS:") {
					return
				}
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
