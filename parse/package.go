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
	NoTestSlice Events

	// Cached indicates whether the test result was obtained from the cache.
	Cached bool

	// Cover reports whether the package contains coverage (go test run with -cover)
	Cover    bool
	Coverage float64

	// HasPanic marks the entire package as panicked. Game over.
	HasPanic bool
}

func NewPackage() *Package {
	return &Package{
		Summary: &Event{},
		Tests:   []*Test{},
	}
}

// AddEvent adds the event to a test based on test name.
func (p *Package) AddEvent(event *Event) {
	for _, t := range p.Tests {
		if t.Name == event.Test {
			t.Events = append(t.Events, event)
			return
		}
	}

	t := &Test{
		Name:    event.Test,
		Package: event.Package,
	}
	t.Events = append(t.Events, event)

	p.Tests = append(p.Tests, t)
}

// TestsByAction returns all tests that identify as one of the following
// actions: pass, skip or fail.
//
// An empty slice if returned if there are no tests.
func (p *Package) TestsByAction(action Action) []*Test {
	tests := []*Test{}

	for _, t := range p.Tests {
		if t.Status() == action {
			tests = append(tests, t)
		}
	}

	return tests
}
