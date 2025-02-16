package parse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ErrNotParsable indicates the event line was not parsable.
var ErrNotParsable = errors.New("failed to parse")

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

	noisy := []string{
		// 1. Filter out noisy output, such as === RUN, === PAUSE, etc.
		updatePrefixRun,
		updatePrefixPause,
		updatePrefixCont,
		updatePrefixPass,
		updatePrefixSkip,
		// 2. Filter out report output, such as --- PASS: and --- SKIP:
		resultPrefixPass,
		resultPrefixSkip,
	}
	isNoisy := func(e *Event) bool {
		output := strings.TrimSpace(e.Output)
		// If the event is a big pass or fail, we can safely discard it. These are typically the
		// lines preceding the package summary line. For example:
		//
		//  PASS
		//  ok      fmt 0.144s
		if e.Test == "" && (output == bigPass || output == bigFail) {
			return true
		}
		for _, prefix := range noisy {
			if strings.HasPrefix(output, prefix) {
				return true
			}
		}
		return false

	}

	sc := bufio.NewScanner(r)
	var started bool
	var badLines int
	for sc.Scan() {
		// Scan up-to 50 lines for a parsable event, if we get one, expect
		// no errors to follow until EOF.
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			// We failed to parse a go test JSON event, but there are special cases for failed
			// builds, setup, etc. Let special case these and bubble them up in the summary
			// if the output belongs to a package.
			summary.AddRawEvent(sc.Text())

			badLines++
			if started || badLines > 50 {
				var syntaxError *json.SyntaxError
				if errors.As(err, &syntaxError) {
					err = fmt.Errorf("line %d JSON error: %s: %w", badLines, syntaxError.Error(), ErrNotParsable)
					if option.debug {
						// In debug mode we can surface a more verbose error message which
						// contains the current line number and exact JSON parsing error.
						fmt.Fprintf(os.Stderr, "debug: %s", err.Error())
					}
				}
				return nil, err
			}
			if option.follow && option.w != nil {
				fmt.Fprintf(option.w, "%s\n", sc.Bytes())
			}
			continue
		}
		started = true

		// TODO(mf): when running tparse locally it's very useful to see progress for long-running
		// test suites. Since we have access to the event we can send it on a chan
		// or just directly update a spinner-like component. This cannot be run with the
		// follow option. Lastly, need to consider what local vs CI behavior would be like.
		// Depending on how often the frames update, this could cause a lot of noise, so maybe
		// we need to expose an interval option, so in CI it would update infrequently.

		// Optionally, as test output is piped to us, we write the plain
		// text Output as if go test was run without the -json flag.
		if (option.follow || option.followVerbose) && option.w != nil {
			if !option.followVerbose && isNoisy(e) {
				continue
			}
			fmt.Fprint(option.w, e.Output)
		}
		// Progress is a special case of follow, where we only print the
		// progress of the test suite, but not the output.
		if option.progress && option.w != nil {
			printProgress(option.progressOutput, e, summary.Packages)
		}

		// TODO(mf): special case build output for now. Need to understand how to better handle this
		// But we don't want to swallow important build errors. There is a class of build output
		// that is bengin like: https://github.com/golang/go/issues/61229
		//
		//  Example:
		//  ld: warning: '.../go.o' has malformed LC_DYSYMTAB, expected 92 undefined symbols to start at index 15983, found 102 undefined symbol
		//
		// TL;DR - output ALL build output to stderr and exclude it from being added to test events
		if e.ImportPath != "" {
			if e.Output != "" {
				fmt.Fprint(os.Stderr, e.Output)
			}
			continue
		}

		summary.AddEvent(e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("received scanning error: %w", err)
	}
	// Entire input has been scanned and no go test JSON output was found.
	if !started {
		return nil, ErrNotParsable
	}

	return summary, nil
}

// printProgress prints a single summary line for each PASS or FAIL package.
// This is useful for long-running test suites.
func printProgress(w io.Writer, e *Event, summary map[string]*Package) {
	if !e.LastLine() {
		return
	}
	action := e.Action
	var suffix string
	if pkg, ok := summary[e.Package]; ok {
		if pkg.NoTests {
			suffix = " [no tests to run]"
			action = ActionSkip
		}
		if pkg.NoTestFiles {
			suffix = " [no test files]"
			action = ActionSkip
		}
	}
	// Normal go test output will print the package summary line like so:
	//
	// FAIL
	// FAIL    github.com/pressly/goose/v4/internal/sqlparser  0.577s
	//
	// PASS
	// ok      github.com/pressly/goose/v4/internal/sqlparser  0.349s
	//
	// ?       github.com/pressly/goose/v4/internal/check      [no test files]
	//
	// testing: warning: no tests to run
	// PASS
	// ok      github.com/pressly/goose/v4/pkg/source  0.382s [no tests to run]
	//
	// We modify this output slightly so it's more consistent and easier to parse.
	fmt.Fprintf(w, "[%s]\t%10s\t%s%s\n",
		strings.ToUpper(action.String()),
		strconv.FormatFloat(e.Elapsed, 'f', 2, 64)+"s",
		e.Package,
		suffix,
	)
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
	// Capture the start time of the package. This is only available in go1.20 and above.
	if e.Action == ActionStart {
		pkg.StartTime = e.Time
		return
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
		// TODO(mf): revisit this behavior?
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
