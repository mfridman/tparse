package parse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
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
// and parses go test output in JSON format until EOF.
//
// Note, Process will attempt to parse up to 50 lines before returning an error.
func Process(r io.Reader, opts ...OptionsFunc) (*GoTestSummary, error) {
	option := &options{}
	for _, f := range opts {
		f(option)
	}
	summary := &GoTestSummary{
		Packages: make(Packages),
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
					err = fmt.Errorf("line %d json error: %s: %w", badLines, err.Error(), ErrNotParseable)
					if option.debug {
						// In debug mode we can surface a more verbose error message which
						// contains the current line number and exact JSON parsing error.
						log.Println("debug: ", err.Error())
					}
					return nil, err
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

		summary.AddEvent(e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("received scanning error: %w", err)
	}
	// Entire input has been scanned and no go test json output was found.
	if !started {
		return nil, ErrNotParseable
	}

	return summary, nil
}

type GoTestSummary struct {
	Packages Packages
}

func (s *GoTestSummary) AddEvent(e *Event) {
	// Discard noisy output such as "=== CONT", "=== RUN", etc. These add
	// no value to the go test output, unless you care to follow how often
	// tests are paused and for what duration.
	if e.Action == ActionOutput && e.DiscardOutput() {
		return
	}
	pkg, ok := s.Packages[e.Package]
	if !ok {
		pkg = newPackage()
		s.Packages[e.Package] = pkg
	}
	// Special case panics.
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
	if e.LastLine() {
		pkg.Summary = e
		return
	}
	// Parse the raw output to add additional metadata to Package.
	switch {
	case e.IsRace():
		pkg.DataRace = append(pkg.DataRace, DataRace{
			PackageName: e.Package,
			TestName:    e.Test,
		})
	case e.IsCached():
		pkg.Cached = true
	case e.NoTestFiles():
		pkg.NoTestFiles = true
		// Manually mark [no test files] as "pass", because the go test tool reports the
		// package Summary action as "skip".
		// TODO(mf): revisit this behaviour?
		pkg.Summary.Package = e.Package
		pkg.Summary.Action = ActionPass
	case e.NoTestsWarn():
		// One or more tests within the package contains no tests.
		pkg.NoTestSlice = append(pkg.NoTestSlice, e)
	case e.NoTestsToRun():
		// Only packages marked as "pass" will contain a summary line appended with [no tests to run].
		// This indicates one or more tests is marked as having no tests to run.
		pkg.NoTests = true
		pkg.Summary.Package = e.Package
		pkg.Summary.Action = ActionPass
	default:
		if cover, ok := e.Cover(); ok {
			pkg.Cover = true
			pkg.Coverage = cover
		}
	}
	pkg.AddEvent(e)
}

func (s *GoTestSummary) GetSortedPackages() []*Package {
	packages := make([]*Package, 0, len(s.Packages))
	for _, pkg := range s.Packages {
		packages = append(packages, pkg)
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Summary.Package < packages[j].Summary.Package
	})
	return packages
}

func (s *GoTestSummary) ExitCode() int {
	for _, pkg := range s.Packages {
		if pkg.HasPanic || len(pkg.DataRace) > 0 || pkg.Summary.Action == ActionFail {
			return 1
		}
	}
	return 0
}
