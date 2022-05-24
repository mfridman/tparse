package parse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mfridman/tparse/internal/check"
)

func TestNewEvent(t *testing.T) {
	t.Parallel()

	tt := []struct {
		raw                    string
		action                 Action
		pkg                    string
		test                   string
		output                 string
		discardOutput          bool
		lastLine               bool
		discardEmptyTestOutput bool
	}{
		{
			// 0
			`{"Time":"2018-10-15T21:03:52.728302-04:00","Action":"run","Package":"fmt","Test":"TestFmtInterface"}`,
			ActionRun, "fmt", "TestFmtInterface", "", false, false, false,
		},
		{
			// 1
			`{"Time":"2018-10-15T21:03:56.232164-04:00","Action":"output","Package":"strings","Test":"ExampleBuilder","Output":"--- PASS: ExampleBuilder (0.00s)\n"}`,
			ActionOutput, "strings", "ExampleBuilder", "--- PASS: ExampleBuilder (0.00s)\n", false, false, false,
		},
		{
			// 2
			`{"Time":"2018-10-15T21:03:56.235807-04:00","Action":"pass","Package":"strings","Elapsed":3.5300000000000002}`,
			ActionPass, "strings", "", "", false, true, false,
		},
		{
			// 3
			`{"Time":"2018-10-15T21:00:51.379156-04:00","Action":"pass","Package":"fmt","Elapsed":0.066}`,
			ActionPass, "fmt", "", "", false, true, false,
		},
		{
			// 4
			`{"Time":"2018-10-15T22:57:28.23799-04:00","Action":"pass","Package":"github.com/astromail/rover/tests","Elapsed":0.582}`,
			ActionPass, "github.com/astromail/rover/tests", "", "", false, true, false,
		},
		{
			// 5
			`{"Time":"2018-10-15T21:00:38.738631-04:00","Action":"pass","Package":"strings","Test":"ExampleTrimRightFunc","Elapsed":0}`,
			ActionPass, "strings", "ExampleTrimRightFunc", "", false, false, false,
		},
		{
			// 6
			`{"Time":"2018-10-15T23:00:27.929094-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/15 23:00:27 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n"}`,
			ActionOutput,
			"github.com/astromail/rover/tests",
			"",
			"2018/10/15 23:00:27 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n",
			false,
			false,
			true,
		},
		{
			// 7
			`{"Time":"2018-10-15T23:00:28.430825-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"PASS\n"}`,
			ActionOutput, "github.com/astromail/rover/tests", "", "PASS\n", false, false, true,
		},
		{
			// 8
			`{"Time":"2018-10-15T23:00:28.432239-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"ok  \tgithub.com/astromail/rover/tests\t0.530s\n"}`,
			ActionOutput,
			"github.com/astromail/rover/tests",
			"",
			"ok  \tgithub.com/astromail/rover/tests\t0.530s\n",
			false,
			false,
			true,
		},
		{
			// 9
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`,
			ActionOutput,
			"github.com/mfridman/srfax",
			"",
			"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n",
			false,
			false,
			true,
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(tc.raw))
			check.NoError(t, err)

			if e.Action != tc.action {
				t.Errorf("wrong action: got %q, want %q", e.Action, tc.action)
			}
			if e.Package != tc.pkg {
				t.Errorf("wrong pkg name: got %q, want %q", e.Package, tc.pkg)
			}
			if e.Output != tc.output {
				t.Errorf("wrong output: got %q, want %q", e.Output, tc.output)
			}
			if e.Test != tc.test {
				t.Errorf("wrong test name: got %q, want %q", e.Test, tc.test)
			}
			if e.LastLine() != tc.lastLine {
				t.Errorf("failed lastLine check: got %v, want %v", e.LastLine(), tc.lastLine)
			}
			if e.DiscardOutput() != tc.discardOutput {
				t.Errorf("failed discard check: got %v, want %v", e.DiscardOutput(), tc.discardOutput)
			}
			if e.DiscardEmptyTestOutput() != tc.discardEmptyTestOutput {
				t.Errorf("failed discard empty test output check: got %v, want %v", e.DiscardEmptyTestOutput(), tc.discardOutput)
			}
			if t.Failed() {
				t.Logf("failed event: %v", tc.raw)
			}
		})
	}
}

func TestCachedEvent(t *testing.T) {
	t.Parallel()

	tt := []struct {
		raw    string
		cached bool
	}{
		{
			// 0
			`{"Time":"2018-10-24T08:30:14.566611-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Output":"ok  \tgithub.com/mfridman/tparse/tests\t(cached)\n"}`,
			true,
		},
		{
			// 1
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`,
			true,
		},
		{
			// 2
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"github.com/mfridman/srfax\t(cached)"}`,
			false,
		},
		{
			// 3
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"(cached)"}`,
			false,
		},
		{
			// 4
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":""}`,
			false,
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(tc.raw))
			check.NoError(t, err)

			got := e.IsCached()
			want := tc.cached
			if got != want {
				t.Errorf("got non-cached output (%t), want cached output (%t)", got, want)
				t.Logf("input: %v", tc.raw)
			}
		})
	}
}
func TestCoverEvent(t *testing.T) {
	t.Parallel()

	var zero float64

	tt := []struct {
		raw      string
		cover    bool
		coverage float64
	}{
		{
			// 0
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`, true, 28.8,
		},
		{
			// 1
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 100.0% of statements\n"}`, true, 100.0,
		},
		{
			// 2
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 0.0% of statements\n"}`, true, zero,
		},
		{
			// 3
			`{"Time":"2018-10-24T09:25:59.855826-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t0.027s\tcoverage: 87.5% of statements\n"}`, true, 87.5,
		},
		{
			// 4
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 1000.0% of statements\n"}`, true, zero,
		},
		{
			// 5
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: .0% of statements\n"}`, false, zero,
		},
		{
			// 6
			`{"Time":"2022-05-23T23:07:54.485803-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Output":"ok  \tgithub.com/mfridman/tparse/tests\t0.516s\tcoverage: 34.5% of statements in ./...\n"}`, true, 34.5,
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(tc.raw))
			check.NoError(t, err)

			f, ok := e.Cover()
			if ok != tc.cover {
				t.Errorf("got (%t) non-coverage event, want %t", ok, tc.cover)
			}
			if f != tc.coverage {
				t.Errorf("got wrong percentage for coverage %v, want %v", f, tc.coverage)
			}
			if t.Failed() {
				t.Logf("input: %v", tc.raw)
			}
		})
	}
}

func TestNoTestFiles(t *testing.T) {
	t.Parallel()
	// [no test files]
	tt := []struct {
		raw         string
		noTestFiles bool
	}{
		{
			// 0
			`{"Time": "2018-10-28T00:06:53.478265-04:00", "Action": "output", "Package": "github.com/astromail/rover", "Output": "?   \tgithub.com/astromail/rover\t[no test files]\n"}`, true,
		},
		{
			// 1
			`{"Time": "2018-10-28T00:06:53.511804-04:00", "Action": "output", "Package": "github.com/astromail/rover/cmd/roverd", "Output": "?   \tgithub.com/astromail/rover/cmd/roverd\t[no test files]\n"}`, true,
		},
		{
			// 2
			`{"Time": "2018-10-28T00:06:53.511804-04:00", "Action": "output", "Package": "github.com/astromail/rover/cmd/roverd", "Output": "   \tgithub.com/astromail/rover/cmd/roverd\t[no test files]"}`, false,
		},
		{
			// 3
			`{"Time": "2018-10-28T00:06:53.511804-04:00", "Action": "output", "Package": "github.com[no test files]\n"}`, false,
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(tc.raw))
			check.NoError(t, err)

			got := e.NoTestFiles()
			want := tc.noTestFiles
			if got != want {
				t.Errorf("got (%t), want (%t) for no test files", got, want)
				t.Logf("input: %v", tc.raw)
			}
		})
	}
}

func TestNoTestsToRun(t *testing.T) {
	t.Parallel()
	// [no test files]
	// This is testing the "package" level
	tt := []struct {
		raw     string
		noTests bool
	}{
		{
			// 0
			`{"Time":"2018-10-28T18:20:47.18358917-04:00","Action":"output","Package":"github.com/awesome/james","Output":"ok  \tgithub.com/awesome/james\t(cached) [no tests to run]\n"}`, true,
		},
		{
			// 1
			`{"Time": "2018-10-28T00:06:53.511804-04:00", "Action": "output", "Package": "github.com/astromail/rover/cmd/roverd", "Output": "?   \tgithub.com/astromail/rover/cmd/roverd\t[no test files]\n"}`, false,
		},
		{
			`{"Time":"2018-10-29T09:31:49.853255-04:00","Action":"output","Package":"github.com/outerspace/v1/tests","Test":"TestSatelliteTransponder","Output":"testing: warning: no tests to run\n"}`, false,
		},
		{
			`{"Time":"2018-10-28T18:20:47.18358917-04:00","Action":"output","Package":"github.com/a/tests","Output":"ok  \tgithub.com/a/tests\t(cached) [no tests to run]\n"}`, true,
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(tc.raw))
			check.NoError(t, err)

			got := e.NoTestsToRun()
			want := tc.noTests
			if got != want {
				t.Errorf("got (%t), want (%t) for no tests to run", got, want)
				t.Logf("input: %v", tc.raw)
			}
		})
	}
}

func TestNoTestsWarn(t *testing.T) {
	t.Parallel()
	// [no test files]
	// This is testing the "test" level only
	tt := []struct {
		raw        string
		wanNoTests bool
	}{
		{
			// 0
			`{"Time":"2018-10-28T18:20:47.18358917-04:00","Action":"output","Package":"github.com/awesome/james","Output":"ok  \tgithub.com/awesome/james\t(cached) [no tests to run]\n"}`, false,
		},
		{
			// 1
			`{"Time": "2018-10-28T00:06:53.511804-04:00", "Action": "output", "Package": "github.com/astromail/rover/cmd/roverd", "Output": "?   \tgithub.com/astromail/rover/cmd/roverd\t[no test files]\n"}`, false,
		},
		{
			// 2
			`{"Time":"2018-10-29T09:31:49.853255-04:00","Action":"output","Package":"github.com/outerspace/v1/tests","Test":"TestSatelliteTransponder","Output":"testing: warning: no tests to run\n"}`, true,
		},
		{
			// 3
			`{"Time":"2018-10-28T18:20:47.245658814-04:00","Action":"output","Package":"github.com/abc/tests","Output":"testing: warning: no tests to run\n"}`, false,
		},
	}

	for i, tc := range tt {
		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(tc.raw))
			check.NoError(t, err)

			got := e.NoTestsWarn()
			want := tc.wanNoTests
			if got != want {
				t.Errorf("got (%t), want (%t) for warn no tests to run", got, want)
				t.Logf("input: %v", tc.raw)
			}
		})
	}
}

func TestActionString(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Action
		want string
	}{
		{ActionRun, "RUN"},
		{ActionPause, "PAUSE"},
		{ActionCont, "CONT"},
		{ActionPass, "PASS"},
		{ActionFail, "FAIL"},
		{ActionOutput, "OUTPUT"},
		{ActionSkip, "SKIP"},
	}
	for _, tc := range tt {
		upper := strings.ToUpper(tc.String())
		if upper != tc.want {
			t.Errorf("got %q, want %q", upper, tc.want)
		}
	}
}

func TestDiscardOutput(t *testing.T) {
	t.Parallel()

	// Table test for JSON events that should be discarded
	tt := []string{
		`{"Time":"2018-11-24T23:18:44.381562-05:00","Action":"output","Package":"time","Test":"TestMonotonicOverflow","Output":"=== RUN   TestMonotonicOverflow\n"}`,

		`{"Time":"2018-10-28T23:41:31.939308-04:00","Action":"output","Package":"github.com/mfridman/tparse/parse","Test":"TestNewEvent","Output":"=== PAUSE TestNewEvent\n"}`,

		`{"Time":"2022-05-20T20:16:06.761846-04:00","Action":"output","Package":"github.com/pressly/goose/v3/tests/e2e","Test":"TestNowAllowMissingUpByOne","Output":"=== CONT  TestNowAllowMissingUpByOne\n"}`,
	}
	for _, tc := range tt {
		e, err := NewEvent([]byte(tc))
		check.NoError(t, err)
		if e.DiscardOutput() != true {
			t.Errorf("%s - %s failed discard check: got:%v, want:%v", e.Package, e.Test, e.DiscardOutput(), true)
		}
	}
}
