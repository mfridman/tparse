package parse

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

func NewEvent(r io.Reader) (*Event, error) {
	var ev Event
	if err := json.NewDecoder(r).Decode(&ev); err != nil {
		return nil, errors.Wrap(err, "failed to parse test event")
	}

	return &ev, nil
}

// Event is a single line of json output from go test with the -json flag.
// For more info see, https://golang.org/cmd/test2json.
type Event struct {
	// Action can be one of:
	// run, pause, cont, pass, bench, fail, output, skip
	Action Action

	// Portion of the test's output (standard output and standard error merged together
	Output string

	// The Time field holds the time the event happened.
	// It is conventionally omitted for cached test results.
	// encodes as an RFC3339-format string
	Time time.Time

	// The Package field, if present, specifies the package being tested.
	// When the go command runs parallel tests in -json mode, events from
	// different tests are interlaced; the Package field allows readers to separate them.
	Package string

	// The Test field, if present, specifies the test, example, or benchmark
	// function that caused the event. Events for the overall package test do not set Test.
	Test string

	// The Elapsed field is set for "pass" and "fail" events.
	// It gives the time elapsed (in seconds) for the specific test or
	// the overall package test that passed or failed.
	Elapsed float64
}

// Events represents all relevant events for a single test.
type Events []*Event

// Discard output-specific events without a test name (with the exception of the final summary line).
func (e *Event) Discard() bool {
	return e.Action == ActionOutput && e.Test == ""
}

// IsSummary checks for the last event summarizing the entire test run. Usually the very
// last line. E.g.,
//
// PASS
// ok  	github.com/astromail/rover/tests	0.583s
// Time:2018-10-14 11:45:03.489687 -0400 EDT Action:pass Output: Package:github.com/astromail/rover/tests Test: Elapsed:0.584
//
// OR
// FAIL
// FAIL	github.com/astromail/rover/tests	0.534s
// Time:2018-10-14 11:45:23.916729 -0400 EDT Action:fail Output: Package:github.com/astromail/rover/tests Test: Elapsed:0.53
func (e *Event) IsSummary() bool {
	return e.Output == "" && e.Test == "" && (e.Action == ActionPass || e.Action == ActionFail)
}

// Action is one of a fixed set of actions describing the event.
type Action string

// Test actions describe a test event. Prefixed with Action for convenience.
const (
	ActionRun    Action = "run"    // test has started running
	ActionPause         = "pause"  // test has been paused
	ActionCont          = "cont"   // the test has continued running
	ActionPass          = "pass"   // test passed
	ActionBench         = "bench"  // benchmark printed log output but did not fail
	ActionFail          = "fail"   // test or benchmark failed
	ActionOutput        = "output" // test printed output
	ActionSkip          = "skip"   // test was skipped or the package contained no tests
)

func (a Action) String() string {
	return strings.ToUpper(string(a))
}

func (a Action) WithColor() string {
	switch a {
	case ActionPass:
		return a.Green()
	case ActionSkip:
		return a.Yellow()
	case ActionFail:
		return a.Red()
	default:
		return a.String()
	}
}

func (a Action) Red() string {
	return color.New(color.FgHiRed).SprintFunc()(a.String())
}

func (a Action) Green() string {
	return color.New(color.FgHiGreen).SprintFunc()(a.String())
}

func (a Action) Yellow() string {
	return color.New(color.FgHiYellow).SprintFunc()(a.String())
}
