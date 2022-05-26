package parse

// Package is the representation of a single package being tested. The
// summary field is an event that contains all relevant information about the
// package, namely Package (name), Elapsed and Action (big pass or fail).
type Package struct {
	Summary *Event
	Tests   []*Test

	// NoTestFiles indicates whether the package contains tests: [no test files]
	// This only occurs at the package level
	NoTestFiles bool

	// NoTests indicates a package contains one or more files with no tests. This doesn't
	// necessarily mean the file is empty or that the package doesn't have any tests.
	// Unfortunately go test marks the package summary with [no tests to run].
	NoTests bool
	// NoTestSlice holds events that contain "testing: warning: no tests to run" and
	// a non-empty test name.
	NoTestSlice []*Event

	// Cached indicates whether the test result was obtained from the cache.
	Cached bool

	// Cover reports whether the package contains coverage (go test run with -cover)
	Cover    bool
	Coverage float64

	// HasPanic marks the entire package as panicked. Game over.
	HasPanic bool
	// Once a package has been marked HasPanic all subsequent events are added to PanicEvents.
	PanicEvents []*Event

	// HasDataRace marks the entire package as having a data race.
	HasDataRace bool
	// DataRaceTests captures an individual test names as having a data race.
	DataRaceTests []string

	// HasFailedBuildOrSetup marks the package as having a failed build or setup.
	// Example: [build failed] or [setup failed]
	HasFailedBuildOrSetup bool
}

// newPackage initializes and returns a Package.
func newPackage() *Package {
	return &Package{
		Summary: &Event{},
		Tests:   []*Test{},
	}
}

// AddEvent adds the event to a test based on test name.
func (p *Package) AddEvent(event *Event) {
	var t *Test
	if t = p.GetTest(event.Test); t == nil {
		// Test does not exist, add it to pkg.
		t = &Test{
			Name:    event.Test,
			Package: event.Package,
		}
		p.Tests = append(p.Tests, t)
	}

	t.Events = append(t.Events, event)
}

// GetTest retuns a test based on given name, if no test is found
// return nil
func (p *Package) GetTest(name string) *Test {
	for _, t := range p.Tests {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// TestsByAction returns all tests that identify as one of the following
// actions: pass, skip or fail.
//
// An empty slice if returned if there are no tests.
func (p *Package) TestsByAction(action Action) []*Test {
	var tests []*Test
	for _, t := range p.Tests {
		if t.Status() == action {
			tests = append(tests, t)
		}
	}
	return tests
}
