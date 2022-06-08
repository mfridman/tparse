package parse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ErrNotParseable indicates the event line was not parseable.
var ErrNotParseable = errors.New("failed to parse")

// Process is the entry point to parse. It consumes a reader
// and parses go test output in JSON format until EOF.
//
// Note, Process will attempt to parse up to 50 lines before returning an error.
func Process(r io.Reader, optionsFunc ...OptionsFunc) (*GoTestSummary, error) {
	option := &options{}
	for _, f := range optionsFunc {
		f(option)
	}
	summary := &GoTestSummary{
		Packages: make(map[string]*Package),
	}

	sc := bufio.NewScanner(r)
	var started bool
	var badLines int
	for sc.Scan() {
		// Scan up-to 50 lines for a parseable event, if we get one, expect
		// no errors to follow until EOF.
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			// We failed to parse a go test JSON event, but there are special cases for failed
			// builds, setup, etc. Let's special case these and bubble them up in the summary
			// if the output belongs to a package.
			summary.AddRawEvent(sc.Text())

			badLines++
			if started || badLines > 50 {
				switch err.(type) {
				case *json.SyntaxError:
					err = fmt.Errorf("line %d json error: %s: %w", badLines, err.Error(), ErrNotParseable)
					if option.debug {
						// In debug mode we can surface a more verbose error message which
						// contains the current line number and exact JSON parsing error.
						fmt.Fprintf(os.Stderr, "debug: %s", err.Error())
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

		// TODO(mf): when running tparse locally it's very useful to see progress for long
		// running test suites. Since we have access to the event we can send it on a chan
		// or just directly update a spinner-like component. This cannot be run with the
		// follow option. Lastly, need to consider what local vs CI behaviour would be like.
		// Depending how often the frames update, this could cause a lot of noise, so maybe
		// we need to expose an interval option, so in CI it would update infrequently.

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
	Packages map[string]*Package
}

func (s *GoTestSummary) AddRawEvent(str string) {
	if strings.HasPrefix(str, "FAIL") {
		ss := failedBuildOrSetupRe.FindStringSubmatch(str)
		if len(ss) == 3 {
			pkgName, failMessage := strings.TrimSpace(ss[1]), strings.TrimSpace(ss[2])
			pkg, ok := s.Packages[pkgName]
			if !ok {
				pkg = newPackage()
				s.Packages[pkgName] = pkg
			}
			pkg.Summary.Package = pkgName
			pkg.Summary.Action = ActionFail
			pkg.Summary.Output = failMessage
			pkg.HasFailedBuildOrSetup = true
		}
	}
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
		pkg.HasDataRace = true
		if e.Test != "" {
			pkg.DataRaceTests = append(pkg.DataRaceTests, e.Test)
		}
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
	// We captured all the necessary package-level information, if the event
	// is output and does not have a test name, discard it.
	if e.DiscardEmptyTestOutput() {
		return
	}
	pkg.AddEvent(e)
}

func (s *GoTestSummary) GetSortedPackages(sorter PackageSorter) []*Package {
	packages := make([]*Package, 0, len(s.Packages))
	for _, pkg := range s.Packages {
		packages = append(packages, pkg)
	}
	sort.Sort(sorter(packages))
	return packages
}

func (s *GoTestSummary) ExitCode() int {
	for _, pkg := range s.Packages {
		switch {
		case pkg.HasFailedBuildOrSetup:
			return 2
		case pkg.HasPanic, pkg.HasDataRace:
			return 1
		case len(pkg.DataRaceTests) > 0:
			return 1
		case pkg.Summary.Action == ActionFail:
			return 1
		}
	}
	return 0
}
