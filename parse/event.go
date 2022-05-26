package parse

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	coverRe              = regexp.MustCompile(`[0-9]{1,3}\.[0-9]{1}\%`)
	failedBuildOrSetupRe = regexp.MustCompile(`^FAIL(.*)\[(build failed|setup failed)\]`)
)

// Event represents a single line of json output from go test with the -json flag.
//
// For more info see, https://golang.org/cmd/test2json and
// https://github.com/golang/go/blob/master/src/cmd/internal/test2json/test2json.go
type Event struct {
	// Action can be one of:
	// run, pause, cont, pass, bench, fail, output, skip
	Action Action

	// Portion of the test's output (standard output and standard error merged together)
	Output string

	// Time at which the the event occurred, encodes as an RFC3339-format string.
	// It is conventionally omitted for cached test results.
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

func (e *Event) String() string {
	return fmt.Sprintf(
		"%-6s - %s - %s elapsed[%.2f] - time[%s]\n%v",
		strings.ToUpper(e.Action.String()),
		e.Package,
		e.Test,
		e.Elapsed,
		e.Time.Format(time.StampMicro),
		e.Output,
	)
}

// NewEvent attempts to decode data into an Event.
func NewEvent(data []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// DiscardOutput reports whether to discard output that belongs to one of
// the output update actions:
// === RUN
// === PAUSE
// === CONT
// If output is none one of the above return false.
func (e *Event) DiscardOutput() bool {
	for i := range updates {
		if strings.HasPrefix(e.Output, updates[i]) {
			return true
		}
	}
	return false
}

func (e *Event) DiscardEmptyTestOutput() bool {
	return e.Action == ActionOutput && e.Test == ""
}

var updates = []string{
	"=== RUN   ",
	"=== PAUSE ",
	"=== CONT  ",
}

// Let's try using the LastLine method to report the package result.
// If there are issues with LastLine() we can switch to this method.
//
// BigResult reports whether the package passed or failed.
// func (e *Event) BigResult() bool {
// 	return e.Test == "" && (e.Output == "PASS\n" || e.Output == "FAIL\n")
// }

// LastLine reports whether the event is the final emitted output line summarizing the package run.
//
// ok  	github.com/astromail/rover/tests	0.583s
// {Time:2018-10-14 11:45:03.489687 -0400 EDT Action:pass Output: Package:github.com/astromail/rover/tests Test: Elapsed:0.584}
//
// FAIL	github.com/astromail/rover/tests	0.534s
// {Time:2018-10-14 11:45:23.916729 -0400 EDT Action:fail Output: Package:github.com/astromail/rover/tests Test: Elapsed:0.53}
func (e *Event) LastLine() bool {
	return e.Test == "" && e.Output == "" && (e.Action == ActionPass || e.Action == ActionFail)
}

// NoTestFiles reports special event case for packages containing no test files:
// "?   \tpackage\t[no test files]\n"
func (e *Event) NoTestFiles() bool {
	return strings.HasPrefix(e.Output, "?   \t") && strings.HasSuffix(e.Output, "[no test files]\n")
}

// NoTestsToRun reports special event case for no tests to run:
// "ok  \tgithub.com/some/awesome/module\t4.543s [no tests to run]\n"
func (e *Event) NoTestsToRun() bool {
	return strings.HasPrefix(e.Output, "ok  \t") && strings.HasSuffix(e.Output, "[no tests to run]\n")
}

// NoTestsWarn whether the event is a test that identifies as: "testing: warning: no tests to run\n"
//
// NOTE: can be found in a package or test event. Must check for non-empty test name in the event.
func (e *Event) NoTestsWarn() bool {
	return e.Test != "" && e.Output == "testing: warning: no tests to run\n"
}

// IsCached reports special event case for cached packages:
// "ok  \tgithub.com/mfridman/tparse/tests\t(cached)\n"
// "ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"
func (e *Event) IsCached() bool {
	return strings.HasPrefix(e.Output, "ok  \t") && strings.Contains(e.Output, "\t(cached)")
}

// Cover reports special event case for package coverage:
// "ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"
// "ok  \tgithub.com/mfridman/srfax\t0.027s\tcoverage: 28.8% of statements\n"
// "ok  \tgithub.com/mfridman/tparse/tests\t0.516s\tcoverage: 34.5% of statements in ./...\n"
func (e *Event) Cover() (float64, bool) {
	var f float64
	var err error
	if strings.Contains(e.Output, "coverage:") && strings.Contains(e.Output, "of statements") {
		s := coverRe.FindString(e.Output)
		f, err = strconv.ParseFloat(strings.TrimRight(s, "%"), 64)
		if err != nil {
			return f, false
		}
		return f, true
	}
	return f, false
}

// IsRace indicates a race event has been detected.
func (e *Event) IsRace() bool {
	return strings.HasPrefix(e.Output, "WARNING: DATA RACE")
}

// IsPanic indicates a panic event has been detected.
func (e *Event) IsPanic() bool {
	// Let's see how this goes. If a user has this in one of their output lines, I think it's
	// defensible to suggest updating their output.
	if strings.HasPrefix(e.Output, "panic: ") {
		return true
	}
	// The golang/go test suite occasionally outputs these keywords along with "as expected":
	// time_test.go:1359: panic in goroutine 7, as expected, with "runtime error: racy use of timers"
	if strings.Contains(e.Output, "runtime error:") && !strings.Contains(e.Output, "as expected") {
		return true
	}
	return false

}

// Action is one of a fixed set of actions describing a single emitted event.
type Action string

// Prefixed with Action for convenience.
const (
	ActionRun    Action = "run"    // test has started running
	ActionPause  Action = "pause"  // test has been paused
	ActionCont   Action = "cont"   // the test has continued running
	ActionPass   Action = "pass"   // test passed
	ActionBench  Action = "bench"  // benchmark printed log output but did not fail
	ActionFail   Action = "fail"   // test or benchmark failed
	ActionOutput Action = "output" // test printed output
	ActionSkip   Action = "skip"   // test was skipped or the package contained no tests
)

func (a Action) String() string {
	return string(a)
}
