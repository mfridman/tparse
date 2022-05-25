package parse

import (
	"sort"
)

// Test represents a single, unique, package test.
type Test struct {
	Name    string
	Package string
	Events  []*Event
}

// Elapsed indicates how long a given test ran (in seconds), by scanning for the largest
// elapsed value from all events.
func (t *Test) Elapsed() float64 {
	var f float64
	for _, e := range t.Events {
		if e.Elapsed > f {
			f = e.Elapsed
		}
	}
	return f
}

// Status reports the outcome of the test represented as a single Action: pass, fail or skip.
func (t *Test) Status() Action {
	// sort by time and scan for an action in reverse order.
	// The first action we come across (in reverse order) is
	// the outcome of the test, which will be one of pass|fail|skip.
	t.SortEvents()
	for i := len(t.Events) - 1; i >= 0; i-- {
		switch t.Events[i].Action {
		case ActionPass:
			return ActionPass
		case ActionSkip:
			return ActionSkip
		case ActionFail:
			return ActionFail
		}
	}
	return ActionFail
}

// SortEvents sorts test events by elapsed time in ascending order, i.e., oldest to newest.
func (t *Test) SortEvents() {
	sort.Slice(t.Events, func(i, j int) bool {
		return t.Events[i].Time.Before(t.Events[j].Time)
	})
}
