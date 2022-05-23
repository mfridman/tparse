package parse

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestNewEvent(t *testing.T) {

	t.Parallel()

	tt := []struct {
		event             string
		action            Action
		pkg               string
		test              string
		output            string
		discard, lastLine bool
	}{
		{
			// 0
			`{"Time":"2018-10-15T21:03:52.728302-04:00","Action":"run","Package":"fmt","Test":"TestFmtInterface"}`,
			ActionRun, "fmt", "TestFmtInterface", "", false, false,
		},
		{
			// 1
			`{"Time":"2018-10-15T21:03:56.232164-04:00","Action":"output","Package":"strings","Test":"ExampleBuilder","Output":"--- PASS: ExampleBuilder (0.00s)\n"}`,
			ActionOutput, "strings", "ExampleBuilder", "--- PASS: ExampleBuilder (0.00s)\n", false, false,
		},
		{
			// 2
			`{"Time":"2018-10-15T21:03:56.235807-04:00","Action":"pass","Package":"strings","Elapsed":3.5300000000000002}`,
			ActionPass, "strings", "", "", false, true,
		},
		{
			// 3
			`{"Time":"2018-10-15T21:00:51.379156-04:00","Action":"pass","Package":"fmt","Elapsed":0.066}`,
			ActionPass, "fmt", "", "", false, true,
		},
		{
			// 4
			`{"Time":"2018-10-15T22:57:28.23799-04:00","Action":"pass","Package":"github.com/astromail/rover/tests","Elapsed":0.582}`,
			ActionPass, "github.com/astromail/rover/tests", "", "", false, true,
		},
		{
			// 5
			`{"Time":"2018-10-15T21:00:38.738631-04:00","Action":"pass","Package":"strings","Test":"ExampleTrimRightFunc","Elapsed":0}`,
			ActionPass, "strings", "ExampleTrimRightFunc", "", false, false,
		},
		{
			// 6
			`{"Time":"2018-10-15T23:00:27.929094-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/15 23:00:27 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n"}`,
			ActionOutput,
			"github.com/astromail/rover/tests",
			"",
			"2018/10/15 23:00:27 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n",
			true,
			false,
		},
		{
			// 7
			`{"Time":"2018-10-15T23:00:28.430825-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"PASS\n"}`,
			ActionOutput, "github.com/astromail/rover/tests", "", "PASS\n", true, false,
		},
		{
			// 8
			`{"Time":"2018-10-15T23:00:28.432239-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"ok  \tgithub.com/astromail/rover/tests\t0.530s\n"}`,
			ActionOutput,
			"github.com/astromail/rover/tests",
			"",
			"ok  \tgithub.com/astromail/rover/tests\t0.530s\n",
			true,
			false,
		},
		{
			// 9
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`,
			ActionOutput,
			"github.com/mfridman/srfax",
			"",
			"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n",
			true,
			false,
		},
	}

	for i, test := range tt {

		i, test := i, test

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {

			t.Parallel()

			e, err := NewEvent([]byte(test.event))
			if err != nil {
				t.Errorf("%s: failed to parse test event:\n%v", test.event, err)
			}

			if e.Action != test.action {
				t.Errorf("wrong action: got %q, want %q", e.Action, test.action)
			}
			if e.Package != test.pkg {
				t.Errorf("wrong pkg name: got %q, want %q", e.Package, test.pkg)
			}
			if e.Output != test.output {
				t.Errorf("wrong output: got %q, want %q", e.Output, test.output)
			}
			if e.Test != test.test {
				t.Errorf("wrong test name: got %q, want %q", e.Test, test.test)
			}
			if e.LastLine() != test.lastLine {
				t.Errorf("failed lastLine check: got %v, want %v", e.LastLine(), test.lastLine)
			}
			if e.DiscardOutput() != test.discard {
				t.Errorf("failed discard check: got %v, want %v", e.DiscardOutput(), test.discard)
			}

			if t.Failed() {
				t.Logf("failed event: %v", test.event)
			}

		})

	}
}

func TestCachedEvent(t *testing.T) {

	t.Parallel()

	tt := []struct {
		input  string
		cached bool
	}{
		{
			// 0
			`{"Time":"2018-10-24T08:30:14.566611-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Output":"ok  \tgithub.com/mfridman/tparse/tests\t(cached)\n"}`, true,
		},
		{
			// 1
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`, true,
		},
		{
			// 2
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"github.com/mfridman/srfax\t(cached)"}`, false,
		},
		{
			// 3
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"(cached)"}`, false,
		},
		{
			// 4
			`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":""}`, false,
		},
	}

	for i, test := range tt {

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(test.input))
			if err != nil {
				t.Fatal(err)
			}

			got := e.IsCached()
			want := test.cached

			if got != want {
				t.Errorf("got non-cached output (%t), want cached output (%t)", got, want)
				t.Logf("input: %v", test.input)
			}
		})

	}
}
func TestCoverEvent(t *testing.T) {

	t.Parallel()

	// long live Golang zero value.
	var zero float64

	tt := []struct {
		input    string
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
	}

	for i, test := range tt {

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(test.input))
			if err != nil {
				t.Fatal(err)
			}

			f, ok := e.Cover()
			if ok != test.cover {
				t.Errorf("got (%t) non-coverage event, want %t", ok, test.cover)
			}

			if f != test.coverage {
				t.Errorf("got wrong percentage for coverage %v, want %v", f, test.coverage)
			}

			if t.Failed() {
				t.Logf("input: %v", test.input)
			}
		})

	}
}

func TestNoTestFiles(t *testing.T) {
	// [no test files]

	t.Parallel()

	tt := []struct {
		input       string
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

	for i, test := range tt {

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(test.input))
			if err != nil {
				t.Fatal(err)
			}

			got := e.NoTestFiles()
			want := test.noTestFiles

			if got != want {
				t.Errorf("got (%t), want (%t) for no test files", got, want)
				t.Logf("input: %v", test.input)
			}
		})

	}
}

func TestNoTestsToRun(t *testing.T) {
	// [no test files]

	// This is testing the "package" level

	t.Parallel()

	tt := []struct {
		input   string
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

	for i, test := range tt {

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(test.input))
			if err != nil {
				t.Fatal(err)
			}

			got := e.NoTestsToRun()
			want := test.noTests

			if got != want {
				t.Errorf("got (%t), want (%t) for no tests to run", got, want)
				t.Logf("input: %v", test.input)
			}
		})

	}
}

func TestNoTestsWarn(t *testing.T) {
	// [no test files]

	// This is testing the "test" level only

	t.Parallel()

	tt := []struct {
		input      string
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

	for i, test := range tt {

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {
			e, err := NewEvent([]byte(test.input))
			if err != nil {
				t.Fatal(err)
			}

			got := e.NoTestsWarn()
			want := test.wanNoTests

			if got != want {
				t.Errorf("got (%t), want (%t) for warn no tests to run", got, want)
				t.Logf("input: %v", test.input)
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

func TestPackageCache(t *testing.T) {

	t.Parallel()

	// This test depends on cached_test.json, which contains the output of 4 std lib packages.
	// go clean -testcache
	// go test strings fmt -json
	// go test strings fmt mime time -json

	// Where bool indicates whether the package is expected to be marked as cached.
	expected := map[string]bool{
		"strings": true,
		"fmt":     true,
		"time":    false,
		"mime":    false,
	}

	f := "./testdata/cached_test.json"
	by, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}

	pkgs, err := Process(bytes.NewReader(by))
	if err != nil {
		t.Fatal(err)
	}

	if len(pkgs) != 4 {
		for n := range pkgs {
			t.Log("got pkg name:", n)
		}
		t.Fatalf("got %d packages, want four packages", len(pkgs))
	}

	for name, pkg := range pkgs {
		t.Run(name, func(t *testing.T) {
			wantCached, ok := expected[name]
			if !ok {
				t.Fatalf("got unexpected package name: %q", name)
			}

			if pkg.Cached != wantCached {
				t.Fatalf("got %t, want package %q to have cached field marked %t", pkg.Cached, name, wantCached)
			}
		})
	}
}

func TestPackageCover(t *testing.T) {

	t.Parallel()

	// This test depends on cover_test.json, which contains the output of 3 std lib packages.
	// go test bytes log sort -json -cover

	// Where bool indicates whether the package is expected to have coverage.
	expected := map[string]float64{
		"log":   68.0,
		"bytes": 86.7,
		"sort":  60.8,
	}

	f := "./testdata/cover_test.json"
	by, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}

	pkgs, err := Process(bytes.NewReader(by))
	if err != nil {
		t.Fatal(err)
	}

	if len(pkgs) != 3 {
		for n := range pkgs {
			t.Log("got pkg name:", n)
		}
		t.Fatalf("got %d packages, want three packages", len(pkgs))
	}

	for name, pkg := range pkgs {
		t.Run(name, func(t *testing.T) {
			wantCover, ok := expected[name]
			if !ok {
				t.Fatalf("got unexpected package name: %q", name)
			}

			if pkg.Coverage != wantCover {
				t.Fatalf("got cover %v, want package %q cover to be %v", pkg.Coverage, name, wantCover)
			}

			var f float64
			if pkg.Coverage > f && !pkg.Cover {
				t.Fatalf("got %v, want package %q to have cover field marked as true when coverage %v>%v", pkg.Cover, name, pkg.Coverage, f)
			}
		})
	}
}
