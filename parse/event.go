package parse

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// Event represents a single line of json output from go test with the -json flag.
//
// For more info see, https://golang.org/cmd/test2json and
// https://github.com/golang/go/blob/master/src/cmd/internal/test2json/test2json.go
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

	// Elapsed is time elapsed (in seconds) for the specific test or
	// the overall package test that passed or failed.
	Elapsed float64
}

func NewEvent(data []byte) (*Event, error) {
	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal test event")
	}

	return &ev, nil
}

// Events groups emitted events by test name. All events must belong to a single test
// and thus a single package.
type Events []*Event

// Discard reports whether an "output" action:
// 1. has test name but is an update action: RUN, PAUSE, CONT.
// 2. has no test name
//
// If output is not one of the above return false.
func (e *Event) Discard() bool {
	u := []string{
		"=== RUN",
		"=== PAUSE",
		"=== CONT",
	}

	for i := range u {
		if strings.HasPrefix(e.Output, u[i]) {
			return true
		}
	}

	return e.Action == ActionOutput && e.Test == ""
}

// Let's try using the Summary method to report the package result.
// If there are issues with Summary we can switch to this method.
//
// BigResult reports whether the package passed or failed.
// func (e *Event) BigResult() bool {
// 	return e.Test == "" && (e.Output == "PASS\n" || e.Output == "FAIL\n")
// }

// Summary reports whether the event is the final emitted output line summarizing the package run.
//
// ok  	github.com/astromail/rover/tests	0.583s
// {Time:2018-10-14 11:45:03.489687 -0400 EDT Action:pass Output: Package:github.com/astromail/rover/tests Test: Elapsed:0.584}
//
// FAIL	github.com/astromail/rover/tests	0.534s
// {Time:2018-10-14 11:45:23.916729 -0400 EDT Action:fail Output: Package:github.com/astromail/rover/tests Test: Elapsed:0.53}
func (e *Event) Summary() bool {
	return e.Test == "" && e.Output == "" && (e.Action == ActionPass || e.Action == ActionFail)
}

// SkipLine reports special event case for packages containing no test files:
// "?   \tpackage\t[no test files]\n"
func (e *Event) SkipLine() bool {
	return strings.HasPrefix(e.Output, "?   \t") && strings.HasSuffix(e.Output, "\t[no test files]\n")
}

// IsCached reports special event case for cached packages:
// ok  \tgithub.com/mfridman/tparse/tests\t(cached)\n
// ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n
func (e *Event) IsCached() bool {
	return strings.HasPrefix(e.Output, "ok  \t") && strings.Contains(e.Output, "\t(cached)")
}

// Cover reports special event case for package coverage:
// ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n
// ok  \tgithub.com/mfridman/srfax\t0.027s\tcoverage: 28.8% of statements\n
func (e Event) Cover() (float64, bool) {
	var re = regexp.MustCompile(`[0-9]{1,3}\.[0-9]{1}\%`)

	var f float64
	var err error

	if strings.Contains(e.Output, "\tcoverage:") && strings.HasSuffix(e.Output, "statements\n") {
		s := re.FindString(e.Output)
		f, err = strconv.ParseFloat(strings.TrimRight(s, "%"), 64)
		if err != nil {
			return f, false
		}

		return f, true
	}

	return f, false
}

// Action is one of a fixed set of actions describing a single emitted test event.
type Action string

// Prefixed with Action for convenience.
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
	return string(a)
}

func (a Action) WithColor() string {
	s := strings.ToUpper(a.String())
	switch a {
	case ActionPass:
		return Green(s)
	case ActionSkip:
		return Yellow(s)
	case ActionFail:
		return Red(s)
	default:
		return s
	}
}

func Red(s string) string {
	return color.New(color.FgHiRed).SprintFunc()(s)
}

func Green(s string) string {
	return color.New(color.FgHiGreen).SprintFunc()(s)
}

func Yellow(s string) string {
	return color.New(color.FgHiYellow).SprintFunc()(s)
}
