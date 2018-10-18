package parse

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Test represents an individual package test.
type Test struct {
	Name    string
	Package string
	Events
}

// Less is used for sorting events based on elapsed time in ascending order, i.e., shortest to longest.
func (t *Test) Less(i, j int) bool { return t.Events[i].Time.Before(t.Events[j].Time) }

// Elapsed returns the largest elapsed value from a test event,
// which represents how long the test ran in seconds.
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
//
// If the test contains no events a failure is returned, this is to prevent a panic.
func (t *Test) Status() Action {
	if len(t.Events) == 0 {
		return ActionFail
	}

	// sort by time and scan for an action in reverse order.
	// The first action we come across (in reverse order) is
	// the outcome of the test, which will be one of pass|fail|skip.

	sort.Slice(t.Events, t.Less)

	for i := len(t.Events) - 1; i >= 0; i-- {
		switch t.Events[i].Action {
		case ActionPass:
			return ActionPass
		case ActionFail:
			return ActionFail
		case ActionSkip:
			return ActionSkip
		}
	}
	// We should never get here. All tests must be flagged with a known action.
	panic(fmt.Sprintf("failed test status check: must be one of pass|fail|skip: %+v\n", t))
}

// Stack returns debugging information from output events for failed or skipped tests.
func (t *Test) Stack() string {

	// TODO: do we need to fail here if len(t.Events) == 0 ?

	// Sort by time and scan for the first output containing the string
	// "--- FAIL" or "--- SKIP"; this event marks the beginning for the "stack".
	// Record it and continue adding all subsequent lines.
	sort.Slice(t.Events, t.Less)

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

		if strings.Contains(e.Output, "--- FAIL") || strings.Contains(e.Output, "--- SKIP") {
			cont = true
			stack.WriteString(e.Output)
		}
	}

	return strings.TrimSpace(stack.String())
}

func (t *Test) PrintFail() {
	sort.Slice(t.Events, t.Less)

	fmt.Printf("%s\t%s\t%s\n%s\n\n",
		t.Status().WithColor(),
		strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
		t.Name,
		t.Stack(),
	)
}

func (t *Test) PrintPass() []string {
	sort.Slice(t.Events, t.Less)

	return []string{
		t.Status().WithColor(),
		strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
		t.Name,
		t.Package,
	}
}
