package parse_test

import (
	"fmt"
	"testing"

	"github.com/mfridman/tparse/parse"
	"github.com/pkg/errors"
)

var events = []struct {
	event            string
	action           parse.Action
	pkg              string
	output           string
	discard, summary bool
}{
	{
		// 0
		`{"Time":"2018-10-15T21:03:52.728302-04:00","Action":"run","Package":"fmt","Test":"TestFmtInterface"}`,
		parse.ActionRun, "fmt", "", false, false,
	},
	{
		// 1
		`{"Time":"2018-10-15T21:03:56.232164-04:00","Action":"output","Package":"strings","Test":"ExampleBuilder","Output":"--- PASS: ExampleBuilder (0.00s)\n"}`,
		parse.ActionOutput, "strings", "--- PASS: ExampleBuilder (0.00s)\n", false, false,
	},
	{
		// 2
		`{"Time":"2018-10-15T21:03:56.235807-04:00","Action":"pass","Package":"strings","Elapsed":3.5300000000000002}`,
		parse.ActionPass, "strings", "", false, true,
	},
	{
		// 3
		`{"Time":"2018-10-15T21:00:51.379156-04:00","Action":"pass","Package":"fmt","Elapsed":0.066}`,
		parse.ActionPass, "fmt", "", false, true,
	},
	{
		// 4
		`{"Time":"2018-10-15T22:57:28.23799-04:00","Action":"pass","Package":"github.com/astromail/rover/tests","Elapsed":0.582}`,
		parse.ActionPass, "github.com/astromail/rover/tests", "", false, true,
	},
	{
		// 5
		`{"Time":"2018-10-15T21:00:38.738631-04:00","Action":"pass","Package":"strings","Test":"ExampleTrimRightFunc","Elapsed":0}`,
		parse.ActionPass, "strings", "", false, false,
	},
	{
		// 6
		`{"Time":"2018-10-15T23:00:27.929094-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/15 23:00:27 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n"}`,
		parse.ActionOutput, "github.com/astromail/rover/tests", "2018/10/15 23:00:27 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n", true, false,
	},
	{
		// 7
		`{"Time":"2018-10-15T23:00:28.430825-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"PASS\n"}`,
		parse.ActionOutput, "github.com/astromail/rover/tests", "PASS\n", true, false,
	},
	{
		// 8
		`{"Time":"2018-10-15T23:00:28.432239-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"ok  \tgithub.com/astromail/rover/tests\t0.530s\n"}`,
		parse.ActionOutput, "github.com/astromail/rover/tests", "ok  \tgithub.com/astromail/rover/tests\t0.530s\n", true, false,
	},
	{
		// 9
		`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`,
		parse.ActionOutput, "github.com/mfridman/srfax", "ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n", true, false,
	},
}

func TestNewEvent(t *testing.T) {

	t.Parallel()

	for i, test := range events {

		i, test := i, test

		t.Run(fmt.Sprintf("event_%d", i), func(t *testing.T) {

			t.Parallel()

			e, err := parse.NewEvent([]byte(test.event))
			if err != nil {
				t.Error(errors.Wrapf(err, "failed to parse test event:\n%v", test.event))
			}

			if e.Action != test.action {
				t.Errorf("wrong action: got %q, want %q", e.Action, test.action)
			}
			if e.Package == test.pkg {
				t.Errorf("wrong pkg: got %q, want %q", e.Package, test.pkg)
			}
			if e.Output != test.output {
				t.Errorf("wrong output: got %q, want %q", e.Output, test.output)
			}
			if e.Summary() == test.summary {
				t.Errorf("failed summary check: got %v, want %v", e.Summary(), test.summary)
			}
			if e.Discard() != test.discard {
				t.Errorf("failed discard check: got %v, want %v", e.Discard(), test.discard)
			}

			if t.Failed() {
				t.Logf("failed event: %v", test.event)
			}

		})

	}
}

func TestCachedPackage(t *testing.T) {

	t.Parallel()

	tt := []struct {
		input    string
		isCached bool
	}{
		{`{"Time":"2018-10-24T08:30:14.566611-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Output":"ok  \tgithub.com/mfridman/tparse/tests\t(cached)\n"}`, true},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`, true},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"github.com/mfridman/srfax\t(cached)"}`, false},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"(cached)"}`, false},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":""}`, false},
	}

	for _, test := range tt {
		e, err := parse.NewEvent([]byte(test.input))
		if err != nil {
			t.Fatal(err)
		}

		got := e.IsCached()
		want := test.isCached

		if got != want {
			t.Errorf("got non-cached output (%t), want cached output (%t)", got, want)
			t.Logf("log: input: %v", test.input)
		}
	}
}
func TestCover(t *testing.T) {

	t.Parallel()

	var zero float64

	tt := []struct {
		input    string
		cover    bool
		coverage float64
	}{
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 28.8% of statements\n"}`, true, 28.8},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 100.0% of statements\n"}`, true, 100.0},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 0.0% of statements\n"}`, true, 0.0},
		{`{"Time":"2018-10-24T09:25:59.855826-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t0.027s\tcoverage: 87.5% of statements\n"}`, true, 87.5},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: 1000.0% of statements\n"}`, true, zero},
		{`{"Time":"2018-10-24T08:48:23.634909-04:00","Action":"output","Package":"github.com/mfridman/srfax","Output":"ok  \tgithub.com/mfridman/srfax\t(cached)\tcoverage: .0% of statements\n"}`, false, zero},
	}

	for _, test := range tt {
		e, err := parse.NewEvent([]byte(test.input))
		if err != nil {
			t.Fatal(err)
		}

		f, ok := e.Cover()
		if ok != test.cover {
			t.Errorf("got (%t) non-coverage event, want %t", ok, test.cover)
		}

		if f != test.coverage {
			t.Errorf("got wrong percentage for coervage %v, want %v", f, test.coverage)
		}

		if t.Failed() {
			t.Logf("log: input: %v", test.input)
		}
	}

}
