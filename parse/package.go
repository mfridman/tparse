package parse

import (
	"strings"
)

// Package is the representation of a single package being tested. The
// summary field is an event that contains all relevant information about the
// package, namely Package (name), Elapsed and Action (big pass or fail).
type Package struct {
	Summary *Event
	Tests   []*Test

	// NoTestFiles indicates whether the package contains tests:
	// "?   \tpackage\t[no test files]\n"
	NoTestFiles bool

	// NoTests indicates whether package contains:
	// "[no tests to run]"
	NoTests bool

	// Cached indicates whether the test result was obtained from the cache.
	Cached bool

	// Cover reports whether the package contains coverage (go test run with -cover)
	Cover    bool
	Coverage float64
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

// HasPanic reports whether a package contains a test that panicked.
// A PanicErr is returned if a test contains a panic.
func (p *Package) HasPanic() error {
	for _, t := range p.Tests {
		for i := range t.Events {
			if strings.HasPrefix(t.Events[i].Output, "panic:") && strings.HasPrefix(t.Events[i+1].Output, "\tpanic:") {
				return &PanicErr{Test: t, Summary: t.Events[len(t.Events)-1]}
			}
		}
	}

	return nil
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
