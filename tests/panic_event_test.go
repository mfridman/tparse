package parse_test

import (
	"strings"
	"testing"

	"github.com/mfridman/tparse/parse"
	"github.com/pkg/errors"
)

func TestPanicEvent(t *testing.T) {

	tt := []string{
		inputPanic1,
		inputPanic2,
		inputPanic3,
	}

	// The input contains a test that panicked, we need to catch this.

	for _, input := range tt {

		_, err := parse.Start(strings.NewReader(input))
		switch err := errors.Cause(err).(type) {
		case *parse.PanicErr:
			continue
		default:
			t.Fatalf("got error %v, want PanicErr", err)
		}
	}

}

const inputPanic1 = `{"Time":"2018-10-21T22:15:24.47322-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
{"Time":"2018-10-21T22:15:24.473515-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"=== RUN   TestStatus\n"}
{"Time":"2018-10-21T22:15:24.473542-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"=== PAUSE TestStatus\n"}
{"Time":"2018-10-21T22:15:24.47355-04:00","Action":"pause","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
{"Time":"2018-10-21T22:15:24.473565-04:00","Action":"cont","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus"}
{"Time":"2018-10-21T22:15:24.473573-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"=== CONT  TestStatus\n"}
{"Time":"2018-10-21T22:15:24.473588-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"--- FAIL: TestStatus (0.00s)\n"}
{"Time":"2018-10-21T22:15:24.47549-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"panic: runtime error: invalid memory address or nil pointer dereference [recovered]\n"}
{"Time":"2018-10-21T22:15:24.475513-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\tpanic: runtime error: invalid memory address or nil pointer dereference\n"}
{"Time":"2018-10-21T22:15:24.475532-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x1112389]\n"}
{"Time":"2018-10-21T22:15:24.47554-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\n"}
{"Time":"2018-10-21T22:15:24.475549-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"goroutine 18 [running]:\n"}
{"Time":"2018-10-21T22:15:24.475559-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"testing.tRunner.func1(0xc0000b6300)\n"}
{"Time":"2018-10-21T22:15:24.475567-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/usr/local/go/src/testing/testing.go:792 +0x387\n"}
{"Time":"2018-10-21T22:15:24.475581-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"panic(0x1137980, 0x1262100)\n"}
{"Time":"2018-10-21T22:15:24.475651-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/usr/local/go/src/runtime/panic.go:513 +0x1b9\n"}
{"Time":"2018-10-21T22:15:24.475682-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"github.com/mfridman/tparse/tests_test.TestStatus.func1(0x116177e, 0xe, 0x1185120, 0xc00006c820, 0x0, 0x0, 0x0, 0xc00002e6c0)\n"}
{"Time":"2018-10-21T22:15:24.475695-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/Users/michael.fridman/go/src/github.com/mfridman/tparse/tests/status_test.go:26 +0x69\n"}
{"Time":"2018-10-21T22:15:24.475749-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"path/filepath.walk(0x116177e, 0xe, 0x1185120, 0xc00006c820, 0xc0000666a0, 0x0, 0x10)\n"}
{"Time":"2018-10-21T22:15:24.475773-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/usr/local/go/src/path/filepath/path.go:362 +0xf6\n"}
{"Time":"2018-10-21T22:15:24.475781-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"path/filepath.Walk(0x116177e, 0xe, 0xc0000666a0, 0x1c338b20, 0xf815f)\n"}
{"Time":"2018-10-21T22:15:24.475788-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/usr/local/go/src/path/filepath/path.go:404 +0x105\n"}
{"Time":"2018-10-21T22:15:24.475798-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"github.com/mfridman/tparse/tests_test.TestStatus(0xc0000b6300)\n"}
{"Time":"2018-10-21T22:15:24.475936-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/Users/michael.fridman/go/src/github.com/mfridman/tparse/tests/status_test.go:19 +0x7e\n"}
{"Time":"2018-10-21T22:15:24.475945-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"testing.tRunner(0xc0000b6300, 0x116ab18)\n"}
{"Time":"2018-10-21T22:15:24.475952-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/usr/local/go/src/testing/testing.go:827 +0xbf\n"}
{"Time":"2018-10-21T22:15:24.475959-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"created by testing.(*T).Run\n"}
{"Time":"2018-10-21T22:15:24.475975-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"\t/usr/local/go/src/testing/testing.go:878 +0x353\n"}
{"Time":"2018-10-21T22:15:24.476216-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Output":"FAIL\tgithub.com/mfridman/tparse/tests\t0.014s\n"}
{"Time":"2018-10-21T22:15:24.476261-04:00","Action":"fail","Package":"github.com/mfridman/tparse/tests","Test":"TestStatus","Elapsed":0.014}`

const inputPanic2 = `{"Time":"2018-10-21T23:42:51.496472-04:00","Action":"output","Package":"github.com/mfridman/tparse","Output":"?   \tgithub.com/mfridman/tparse\t[no test files]\n"}
{"Time":"2018-10-21T23:42:51.496734-04:00","Action":"skip","Package":"github.com/mfridman/tparse","Elapsed":0}
{"Time":"2018-10-21T23:42:51.49677-04:00","Action":"output","Package":"github.com/mfridman/tparse/ignore","Output":"?   \tgithub.com/mfridman/tparse/ignore\t[no test files]\n"}
{"Time":"2018-10-21T23:42:51.496782-04:00","Action":"skip","Package":"github.com/mfridman/tparse/ignore","Elapsed":0}
{"Time":"2018-10-21T23:42:51.496805-04:00","Action":"output","Package":"github.com/mfridman/tparse/parse","Output":"?   \tgithub.com/mfridman/tparse/parse\t[no test files]\n"}
{"Time":"2018-10-21T23:42:51.496813-04:00","Action":"skip","Package":"github.com/mfridman/tparse/parse","Elapsed":0}
{"Time":"2018-10-21T23:42:51.673696-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent"}
{"Time":"2018-10-21T23:42:51.673742-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent","Output":"=== RUN   TestNewEvent\n"}
{"Time":"2018-10-21T23:42:51.673772-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent","Output":"=== PAUSE TestNewEvent\n"}
{"Time":"2018-10-21T23:42:51.673795-04:00","Action":"pause","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent"}
{"Time":"2018-10-21T23:42:51.673815-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent"}
{"Time":"2018-10-21T23:42:51.673834-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent","Output":"=== RUN   TestPanicEvent\n"}
{"Time":"2018-10-21T23:42:51.674274-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent","Output":"--- PASS: TestPanicEvent (0.00s)\n"}
{"Time":"2018-10-21T23:42:51.674295-04:00","Action":"pass","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent","Elapsed":0}
{"Time":"2018-10-21T23:42:51.674307-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestStack"}
{"Time":"2018-10-21T23:42:51.674314-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"=== RUN   TestStack\n"}
{"Time":"2018-10-21T23:42:51.674328-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"--- FAIL: TestStack (0.00s)\n"}
{"Time":"2018-10-21T23:42:51.676397-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"panic: oops [recovered]\n"}
{"Time":"2018-10-21T23:42:51.676427-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\tpanic: oops\n"}
{"Time":"2018-10-21T23:42:51.676437-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\n"}
{"Time":"2018-10-21T23:42:51.676453-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"goroutine 20 [running]:\n"}
{"Time":"2018-10-21T23:42:51.676462-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"testing.tRunner.func1(0xc0000b4600)\n"}
{"Time":"2018-10-21T23:42:51.676489-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\t/usr/local/go/src/testing/testing.go:792 +0x387\n"}
{"Time":"2018-10-21T23:42:51.676501-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"panic(0x112ad60, 0x1182f90)\n"}
{"Time":"2018-10-21T23:42:51.67651-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\t/usr/local/go/src/runtime/panic.go:513 +0x1b9\n"}
{"Time":"2018-10-21T23:42:51.676523-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"github.com/mfridman/tparse/tests_test.TestStack(0xc0000b4600)\n"}
{"Time":"2018-10-21T23:42:51.676542-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\t/Users/michael.fridman/go/src/github.com/mfridman/tparse/tests/stack_test.go:12 +0x39\n"}
{"Time":"2018-10-21T23:42:51.676555-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"testing.tRunner(0xc0000b4600, 0x116a730)\n"}
{"Time":"2018-10-21T23:42:51.676585-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\t/usr/local/go/src/testing/testing.go:827 +0xbf\n"}
{"Time":"2018-10-21T23:42:51.676596-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"created by testing.(*T).Run\n"}
{"Time":"2018-10-21T23:42:51.67666-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"\t/usr/local/go/src/testing/testing.go:878 +0x353\n"}
{"Time":"2018-10-21T23:42:51.676943-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Output":"FAIL\tgithub.com/mfridman/tparse/tests\t0.016s\n"}
{"Time":"2018-10-21T23:42:51.676966-04:00","Action":"fail","Package":"github.com/mfridman/tparse/tests","Test":"TestStack","Elapsed":0.016}`

const inputPanic3 = `{"Time":"2018-10-24T14:32:01.310248-04:00","Action":"output","Package":"github.com/mfridman/tparse","Output":"?   \tgithub.com/mfridman/tparse\t[no test files]\n"}
{"Time":"2018-10-24T14:32:01.31052-04:00","Action":"skip","Package":"github.com/mfridman/tparse","Elapsed":0}
{"Time":"2018-10-24T14:32:01.31056-04:00","Action":"output","Package":"github.com/mfridman/tparse/ignore","Output":"?   \tgithub.com/mfridman/tparse/ignore\t[no test files]\n"}
{"Time":"2018-10-24T14:32:01.310588-04:00","Action":"skip","Package":"github.com/mfridman/tparse/ignore","Elapsed":0}
{"Time":"2018-10-24T14:32:01.310626-04:00","Action":"output","Package":"github.com/mfridman/tparse/parse","Output":"?   \tgithub.com/mfridman/tparse/parse\t[no test files]\n"}
{"Time":"2018-10-24T14:32:01.310638-04:00","Action":"skip","Package":"github.com/mfridman/tparse/parse","Elapsed":0}
{"Time":"2018-10-24T14:32:01.499657-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent"}
{"Time":"2018-10-24T14:32:01.499717-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent","Output":"=== RUN   TestNewEvent\n"}
{"Time":"2018-10-24T14:32:01.499741-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent","Output":"=== PAUSE TestNewEvent\n"}
{"Time":"2018-10-24T14:32:01.499758-04:00","Action":"pause","Package":"github.com/mfridman/tparse/tests","Test":"TestNewEvent"}
{"Time":"2018-10-24T14:32:01.499771-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestCachedPackage"}
{"Time":"2018-10-24T14:32:01.499781-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestCachedPackage","Output":"=== RUN   TestCachedPackage\n"}
{"Time":"2018-10-24T14:32:01.499789-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestCachedPackage","Output":"=== PAUSE TestCachedPackage\n"}
{"Time":"2018-10-24T14:32:01.499799-04:00","Action":"pause","Package":"github.com/mfridman/tparse/tests","Test":"TestCachedPackage"}
{"Time":"2018-10-24T14:32:01.499809-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestCover"}
{"Time":"2018-10-24T14:32:01.499816-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestCover","Output":"=== RUN   TestCover\n"}
{"Time":"2018-10-24T14:32:01.499829-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestCover","Output":"=== PAUSE TestCover\n"}
{"Time":"2018-10-24T14:32:01.499836-04:00","Action":"pause","Package":"github.com/mfridman/tparse/tests","Test":"TestCover"}
{"Time":"2018-10-24T14:32:01.499843-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent"}
{"Time":"2018-10-24T14:32:01.499849-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent","Output":"=== RUN   TestPanicEvent\n"}
{"Time":"2018-10-24T14:32:01.500824-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent","Output":"--- PASS: TestPanicEvent (0.00s)\n"}
{"Time":"2018-10-24T14:32:01.500848-04:00","Action":"pass","Package":"github.com/mfridman/tparse/tests","Test":"TestPanicEvent","Elapsed":0}
{"Time":"2018-10-24T14:32:01.500859-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan"}
{"Time":"2018-10-24T14:32:01.500869-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan","Output":"=== RUN   TestPrescan\n"}
{"Time":"2018-10-24T14:32:01.50088-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/pass"}
{"Time":"2018-10-24T14:32:01.50089-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/pass","Output":"=== RUN   TestPrescan/pass\n"}
{"Time":"2018-10-24T14:32:01.501224-04:00","Action":"run","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01"}
{"Time":"2018-10-24T14:32:01.501234-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"=== RUN   TestPrescan/fail01\n"}
{"Time":"2018-10-24T14:32:01.503535-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"panic: runtime error: invalid memory address or nil pointer dereference [recovered]\n"}
{"Time":"2018-10-24T14:32:01.503561-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\tpanic: runtime error: invalid memory address or nil pointer dereference\n"}
{"Time":"2018-10-24T14:32:01.503574-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"[signal SIGSEGV: segmentation violation code=0x1 addr=0x40 pc=0x110f2f1]\n"}
{"Time":"2018-10-24T14:32:01.503582-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\n"}
{"Time":"2018-10-24T14:32:01.503592-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"goroutine 11 [running]:\n"}
{"Time":"2018-10-24T14:32:01.5036-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"testing.tRunner.func1(0xc0000a0b00)\n"}
{"Time":"2018-10-24T14:32:01.503612-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\t/usr/local/go/src/testing/testing.go:792 +0x387\n"}
{"Time":"2018-10-24T14:32:01.503664-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"panic(0x1138e80, 0x126a100)\n"}
{"Time":"2018-10-24T14:32:01.50368-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\t/usr/local/go/src/runtime/panic.go:513 +0x1b9\n"}
{"Time":"2018-10-24T14:32:01.503689-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"github.com/mfridman/tparse/parse.Start(0x118b220, 0xc0000fb8a0, 0x5bd0baa1, 0xc000034798, 0x106e0b6)\n"}
{"Time":"2018-10-24T14:32:01.503697-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\t/Users/michael.fridman/go/src/github.com/mfridman/tparse/parse/package.go:136 +0x151\n"}
{"Time":"2018-10-24T14:32:01.503707-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"github.com/mfridman/tparse/tests_test.TestPrescan.func2(0xc0000a0b00)\n"}
{"Time":"2018-10-24T14:32:01.503779-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\t/Users/michael.fridman/go/src/github.com/mfridman/tparse/tests/prescan_test.go:144 +0x7a\n"}
{"Time":"2018-10-24T14:32:01.503791-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"testing.tRunner(0xc0000a0b00, 0x1171820)\n"}
{"Time":"2018-10-24T14:32:01.503798-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\t/usr/local/go/src/testing/testing.go:827 +0xbf\n"}
{"Time":"2018-10-24T14:32:01.503805-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"created by testing.(*T).Run\n"}
{"Time":"2018-10-24T14:32:01.503813-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"\t/usr/local/go/src/testing/testing.go:878 +0x353\n"}
{"Time":"2018-10-24T14:32:01.504138-04:00","Action":"output","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Output":"FAIL\tgithub.com/mfridman/tparse/tests\t0.016s\n"}
{"Time":"2018-10-24T14:32:01.50418-04:00","Action":"fail","Package":"github.com/mfridman/tparse/tests","Test":"TestPrescan/fail01","Elapsed":0.016}`
