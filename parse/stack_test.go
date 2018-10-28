package parse

import (
	"strings"
	"testing"
)

func TestStack(t *testing.T) {

	colors = false

	t.Parallel()

	// Inputs and outputs only check a single test case.
	// All inputs must have the same test name, otherwise a panic (intentionally) might occur.
	tests := []struct {
		name, input, output string
	}{
		{"input1", stackInput1, stackOutput1},
		{"input2", stackInput2, stackOutput2},
		{"input3", stackInput3, stackOutput3},
	}

	for _, tt := range tests {

		pkgs, err := Start(strings.NewReader(tt.input))
		if err != nil {
			t.Fatalf("failed to get packages: %v", err)
		}

		for _, pkg := range pkgs {
			for _, test := range pkg.Tests {
				if test.Status() != ActionSkip || test.Status() != ActionFail {
					continue
				}

				t.Run(tt.name, func(t *testing.T) {

					got := test.Stack()
					want := tt.output

					if strings.Compare(got, want) != 0 {
						t.Errorf("failed stack comparison")
					}

					if t.Failed() {
						t.Logf("log:\ngot:%v\nwant:%v", got, want)
					}

				}) // end t.Run
			} // done test
		} // done pkg
	}

}

// TODO: consider moving this to a file? As tests are added, this will almost certainly lead to
// confusion on how Go handles string literals: https://golang.org/ref/spec#String_literals.
// Then there will be issues with inputs using spaces and the expected outputs will be formatted with tabs, ugh.

const stackInput1 = `{"Time":"2018-10-17T22:27:03.033477-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
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

const stackOutput1 = `--- FAIL: TestStatus (0.00s)
    status_test.go:91: got no failed tests, want one or more tests to be marked as "fail"`

const stackInput2 = `{"Time":"2018-10-17T22:49:13.689691-04:00","Action":"run","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve"}
{"Time":"2018-10-17T22:49:13.689709-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"=== RUN   TestCatch/catchAndRetrieve\n"}
{"Time":"2018-10-17T22:49:13.694812-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"    --- FAIL: TestCatch/catchAndRetrieve (0.01s)\n"}
{"Time":"2018-10-17T22:49:13.694844-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"        catch_test.go:37: email id does not match: got \"2502e1fe-6d40-42ae-a21a-89b52a1be84c\", want \"123@example.com\"\n"}
{"Time":"2018-10-17T22:49:13.695157-04:00","Action":"fail","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Elapsed":0.01}`

const stackOutput2 = `--- FAIL: TestCatch/catchAndRetrieve (0.01s)
        catch_test.go:37: email id does not match: got "2502e1fe-6d40-42ae-a21a-89b52a1be84c", want "123@example.com"`

const stackInput3 = `{"Time":"2018-10-28T00:06:53.478265-04:00","Action":"output","Package":"github.com/astromail/rover","Output":"?   \tgithub.com/astromail/rover\t[no test files]\n"}
{"Time":"2018-10-28T00:06:53.478512-04:00","Action":"skip","Package":"github.com/astromail/rover","Elapsed":0}
{"Time":"2018-10-28T00:06:53.511804-04:00","Action":"output","Package":"github.com/astromail/rover/cmd/roverd","Output":"?   \tgithub.com/astromail/rover/cmd/roverd\t[no test files]\n"}
{"Time":"2018-10-28T00:06:53.511862-04:00","Action":"skip","Package":"github.com/astromail/rover/cmd/roverd","Elapsed":0}
{"Time":"2018-10-28T00:06:53.511882-04:00","Action":"output","Package":"github.com/astromail/rover/errors","Output":"?   \tgithub.com/astromail/rover/errors\t[no test files]\n"}
{"Time":"2018-10-28T00:06:53.511891-04:00","Action":"skip","Package":"github.com/astromail/rover/errors","Elapsed":0}
{"Time":"2018-10-28T00:06:53.511907-04:00","Action":"output","Package":"github.com/astromail/rover/smtp","Output":"?   \tgithub.com/astromail/rover/smtp\t[no test files]\n"}
{"Time":"2018-10-28T00:06:53.511916-04:00","Action":"skip","Package":"github.com/astromail/rover/smtp","Elapsed":0}
{"Time":"2018-10-28T00:06:53.511933-04:00","Action":"output","Package":"github.com/astromail/rover/storage","Output":"?   \tgithub.com/astromail/rover/storage\t[no test files]\n"}
{"Time":"2018-10-28T00:06:53.511942-04:00","Action":"skip","Package":"github.com/astromail/rover/storage","Elapsed":0}
{"Time":"2018-10-28T00:06:53.511957-04:00","Action":"output","Package":"github.com/astromail/rover/storage/badger","Output":"?   \tgithub.com/astromail/rover/storage/badger\t[no test files]\n"}
{"Time":"2018-10-28T00:06:53.511969-04:00","Action":"skip","Package":"github.com/astromail/rover/storage/badger","Elapsed":0}
{"Time":"2018-10-28T00:06:54.007207-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/28 00:06:54 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n"}
{"Time":"2018-10-28T00:06:54.007282-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/28 00:06:54 Iterating file id: 0\n"}
{"Time":"2018-10-28T00:06:54.007321-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/28 00:06:54 Replaying from value pointer: {Fid:0 Len:0 Offset:0}\n"}
{"Time":"2018-10-28T00:06:54.00733-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/28 00:06:54 Iterating file id: 0\n"}
{"Time":"2018-10-28T00:06:54.007389-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/28 00:06:54 Iteration took: 144.649µs\n"}
{"Time":"2018-10-28T00:06:54.007462-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"2018/10/28 00:06:54 Iteration took: 259.316µs\n"}
{"Time":"2018-10-28T00:06:54.503762-04:00","Action":"run","Package":"github.com/astromail/rover/tests","Test":"TestCatch"}
{"Time":"2018-10-28T00:06:54.503795-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch","Output":"=== RUN   TestCatch\n"}
{"Time":"2018-10-28T00:06:54.503829-04:00","Action":"run","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve"}
{"Time":"2018-10-28T00:06:54.503838-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"=== RUN   TestCatch/catchAndRetrieve\n"}
{"Time":"2018-10-28T00:06:54.507445-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch","Output":"--- FAIL: TestCatch (0.00s)\n"}
{"Time":"2018-10-28T00:06:54.507485-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"    --- FAIL: TestCatch/catchAndRetrieve (0.00s)\n"}
{"Time":"2018-10-28T00:06:54.507507-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"        catch_test.go:29: got id \"ad0892h\", want empty id\n"}
{"Time":"2018-10-28T00:06:54.507517-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"        catch_test.go:37: email id does not match: got \"69c47b65-0ad5-47d5-9346-f4f8bd22c56e\", want \"oops@example.com\"\n"}
{"Time":"2018-10-28T00:06:54.507528-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Output":"        catch_test.go:41: failed to mark email as read: failed to mark email id \"123\" as read: PUT \"http://localhost:8026/api/v1/emails/123/read\": expecting valid status code, got 500 Internal Server Error\n"}
{"Time":"2018-10-28T00:06:54.507536-04:00","Action":"fail","Package":"github.com/astromail/rover/tests","Test":"TestCatch/catchAndRetrieve","Elapsed":0}
{"Time":"2018-10-28T00:06:54.507544-04:00","Action":"fail","Package":"github.com/astromail/rover/tests","Test":"TestCatch","Elapsed":0}
{"Time":"2018-10-28T00:06:54.507549-04:00","Action":"run","Package":"github.com/astromail/rover/tests","Test":"TestNameFormat"}
{"Time":"2018-10-28T00:06:54.507554-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestNameFormat","Output":"=== RUN   TestNameFormat\n"}
{"Time":"2018-10-28T00:06:54.507571-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Test":"TestNameFormat","Output":"--- PASS: TestNameFormat (0.00s)\n"}
{"Time":"2018-10-28T00:06:54.507578-04:00","Action":"pass","Package":"github.com/astromail/rover/tests","Test":"TestNameFormat","Elapsed":0}
{"Time":"2018-10-28T00:06:54.507583-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"FAIL\n"}
{"Time":"2018-10-28T00:06:54.50967-04:00","Action":"output","Package":"github.com/astromail/rover/tests","Output":"FAIL\tgithub.com/astromail/rover/tests\t0.532s\n"}
{"Time":"2018-10-28T00:06:54.509705-04:00","Action":"fail","Package":"github.com/astromail/rover/tests","Elapsed":0.532}`

const stackOutput3 = `--- FAIL: TestCatch (0.00s)--- FAIL: TestCatch/catchAndRetrieve (0.00s)
        catch_test.go:29: got id "ad0892h", want empty id
        catch_test.go:37: email id does not match: got "25a44926-169d-4b45-9673-15d580e697df", want "oops@example.com"
        catch_test.go:41: failed to mark email as read: failed to mark email id "123" as read: PUT "http://localhost:8026/api/v1/emails/123/read": expecting valid status code, got 500 Internal Server Error`
