package parse

import (
	"sort"
	"strings"
)

// Test represents a single, unique, package test.
type Test struct {
	Name    string
	Package string
	Events
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
	t.Sort()

	for i := len(t.Events) - 1; i >= 0; i-- {
		switch t.Events[i].Action {
		case ActionPass:
			return ActionPass
		case ActionSkip:
			return ActionSkip
		}
	}

	return ActionFail
}

// Stack returns debugging information from output events for failed or skipped tests.
func (t *Test) Stack() string {

	// Sort by time and scan for the first output containing the string
	// "--- FAIL" or "--- SKIP"; this event marks the beginning for the "stack".
	// Record it and continue adding all subsequent lines.
	t.Sort()

	ss := []string{
		"--- FAIL:",
		"--- PASS:",
		"--- SKIP:",
		"--- BENCH:",
	}

	var stack strings.Builder

	var cont bool
	for _, e := range t.Events {
		// Only output events have useful information. Skip everything else.
		if e.Action != ActionOutput {
			continue
		}

		if cont {
			stack.WriteString(e.Output)
			continue
		}

		// this is a special case for truly end of the world runtime panics
		if strings.HasPrefix(e.Output, "panic:") && strings.Contains(e.Output, "runtime error:") {
			stack.WriteString(e.Output)
			cont = true
			continue
		}

		for i := range ss {
			if strings.Contains(e.Output, ss[i]) {
				cont = true
				stack.WriteString(colorize(e.Output, cRed, true))
			}
		}
	}

	return strings.TrimSpace(stack.String())
}

// Sort events by elapsed time in ascending order, i.e., oldest to newest.
func (t *Test) Sort() {
	sort.Slice(t.Events, func(i, j int) bool {
		return t.Events[i].Time.Before(t.Events[j].Time)
	})
}
