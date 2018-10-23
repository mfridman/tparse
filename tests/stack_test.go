package parse_test

import (
	"strings"
	"testing"

	"github.com/mfridman/tparse/parse"
)

func TestStack(t *testing.T) {
	t.Parallel()

	// Inputs and outputs only check a single test case.
	// All inputs must have the same test name, otherwise a panic (intentionally) might occur.
	tests := []struct {
		name, input, output string
	}{
		{"input1", input1, output1},
		{"input2", input2, output2},
	}

	for _, tt := range tests {

		pkgs, err := parse.Start(strings.NewReader(tt.input))
		if err != nil {
			t.Fatalf("failed to get packages: %v", err)
		}

		for _, pkg := range pkgs {

			for _, test := range pkg.Tests {

				if test.Status() != parse.ActionFail || test.Status() != parse.ActionSkip {
					continue
				}

				t.Run(test.Name, func(t *testing.T) {

					got := test.Stack()
					want := tt.output

					if strings.Compare(got, want) != 0 {
						t.Errorf("failed stack comparison")
					}

					if t.Failed() {
						t.Logf("log:\ngot:%v\n\nwant:%v", got, want)
					}

				}) // end t.Run
			} // done test
		} // done pkg
	}

}

// TODO: consider moving this to a file? As tests are added, this will almost certainly lead to
// confusion on how Go handles string literals: https://golang.org/ref/spec#String_literals.
// Then there will be issues with inputs using spaces and the expected outputs will be formatted with tabs, ugh.

const input1 = `{"Time":"2018-10-17T22:27:03.033477-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
{"Time":"2018-10-17T22:27:03.033807-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"=== RUN   TestStatus\n"}
{"Time":"2018-10-17T22:27:03.033839-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"=== PAUSE TestStatus\n"}
{"Time":"2018-10-17T22:27:03.033848-04:00","Action":"pause","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
{"Time":"2018-10-17T22:27:03.033862-04:00","Action":"cont","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
{"Time":"2018-10-17T22:27:03.03387-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"=== CONT  TestStatus\n"}
{"Time":"2018-10-17T22:27:03.034043-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"--- FAIL: TestStatus (0.00s)\n"}
{"Time":"2018-10-17T22:27:03.034072-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"    status_test.go:91: got no failed tests, want one or more tests to be marked as \"fail\"\n"}
{"Time":"2018-10-17T22:27:03.034098-04:00","Action":"fail","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Elapsed":0}
{"Time":"2018-10-17T22:27:03.034118-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Output":"FAIL\n"}
{"Time":"2018-10-17T22:27:03.03447-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Output":"FAIL\tgithub.com/mfridman/tparse/tests\t0.012s\n"}`

const output1 = `--- FAIL: TestStatus (0.00s)
    status_test.go:91: got no failed tests, want one or more tests to be marked as "fail"`

const input2 = `{"Time":"2018-10-17T22:49:13.689691-04:00","Action":"run","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve"}
{"Time":"2018-10-17T22:49:13.689709-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"=== RUN   TestCatch/catchAndRetrieve\n"}
{"Time":"2018-10-17T22:49:13.694812-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"    --- FAIL: TestCatch/catchAndRetrieve (0.01s)\n"}
{"Time":"2018-10-17T22:49:13.694844-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"        catch_test.go:37: email id does not match: got \"2502e1fe-6d40-42ae-a21a-89b52a1be84c\", want \"123@example.com\"\n"}
{"Time":"2018-10-17T22:49:13.695157-04:00","Action":"fail","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Elapsed":0.01}`

const output2 = `--- FAIL: TestCatch/catchAndRetrieve (0.01s)
    catch_test.go:37: email id does not match: got "2502e1fe-6d40-42ae-a21a-89b52a1be84c", want "123@example.com"`
