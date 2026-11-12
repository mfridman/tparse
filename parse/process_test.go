package parse

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveredPanicPassIssue145(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		`{"Time":"2025-09-09T11:21:23.522564-07:00","Action":"start","Package":"go.temporal.io/server/service/history/replication"}`,
		`{"Time":"2025-09-09T11:21:23.522664-07:00","Action":"run","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite"}`,
		`{"Time":"2025-09-09T11:21:23.522685-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite","Output":"=== RUN   TestStreamReceiverSuite\n"}`,
		`{"Time":"2025-09-09T11:21:23.522702-07:00","Action":"run","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured"}`,
		`{"Time":"2025-09-09T11:21:23.522707-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured","Output":"=== RUN   TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured\n"}`,
		`{"Time":"2025-09-09T11:21:23.522722-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured","Output":"2025-09-09T11:21:01.477-0700\terror\tPanic is captured\t{\"error\": \"runtime error: invalid memory address or nil pointer dereference\"}\n"}`,
		`{"Time":"2025-09-09T11:21:23.522891-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured","Output":"runtime.gopanic\n"}`,
		`{"Time":"2025-09-09T11:21:23.522897-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured","Output":"\t/opt/homebrew/Cellar/go/1.25.0/libexec/src/runtime/panic.go:783\n"}`,
		`{"Time":"2025-09-09T11:21:23.523124-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite","Output":"--- PASS: TestStreamReceiverSuite (0.00s)\n"}`,
		`{"Time":"2025-09-09T11:21:23.523131-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured","Output":"    --- PASS: TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured (0.00s)\n"}`,
		`{"Time":"2025-09-09T11:21:23.523134-07:00","Action":"pass","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured","Elapsed":0}`,
		`{"Time":"2025-09-09T11:21:23.523138-07:00","Action":"pass","Package":"go.temporal.io/server/service/history/replication","Test":"TestStreamReceiverSuite","Elapsed":0}`,
		`{"Time":"2025-09-09T11:21:23.52314-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Output":"PASS\n"}`,
		`{"Time":"2025-09-09T11:21:23.523146-07:00","Action":"output","Package":"go.temporal.io/server/service/history/replication","Output":"ok  \tgo.temporal.io/server/service/history/replication\t(cached)\n"}`,
		`{"Time":"2025-09-09T11:21:23.523157-07:00","Action":"pass","Package":"go.temporal.io/server/service/history/replication","Elapsed":0.001}`,
	}, "\n")

	summary, err := Process(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, 0, summary.ExitCode())

	pkg, ok := summary.Packages["go.temporal.io/server/service/history/replication"]
	require.True(t, ok)
	assert.False(t, pkg.HasPanic)
	assert.Equal(t, ActionPass, pkg.Summary.Action)
	assert.Empty(t, pkg.PanicEvents)

	subTest := pkg.GetTest("TestStreamReceiverSuite/TestSendEventLoop_Panic_Captured")
	require.NotNil(t, subTest)
	assert.Equal(t, ActionPass, subTest.Status())

	suiteTest := pkg.GetTest("TestStreamReceiverSuite")
	require.NotNil(t, suiteTest)
	assert.Equal(t, ActionPass, suiteTest.Status())

	assert.Len(t, pkg.TestsByAction(ActionPass), 2)
	assert.Empty(t, pkg.TestsByAction(ActionFail))
}

func TestRecoveredPanicStdlib(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		`{"Time":"2025-09-09T11:00:00.000000-07:00","Action":"run","Package":"example.com/pkg","Test":"TestRecovered"}`,
		`{"Time":"2025-09-09T11:00:00.000100-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"=== RUN   TestRecovered\n"}`,
		`{"Time":"2025-09-09T11:00:00.000200-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"panic: expected panic occurred [recovered]\n"}`,
		`{"Time":"2025-09-09T11:00:00.000300-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"--- PASS: TestRecovered (0.00s)\n"}`,
		`{"Time":"2025-09-09T11:00:00.000400-07:00","Action":"pass","Package":"example.com/pkg","Test":"TestRecovered","Elapsed":0}`,
		`{"Time":"2025-09-09T11:00:00.000500-07:00","Action":"output","Package":"example.com/pkg","Output":"PASS\n"}`,
		`{"Time":"2025-09-09T11:00:00.000600-07:00","Action":"output","Package":"example.com/pkg","Output":"ok  \texample.com/pkg\t0.010s\n"}`,
		`{"Time":"2025-09-09T11:00:00.000700-07:00","Action":"pass","Package":"example.com/pkg","Elapsed":0.01}`,
	}, "\n")

	summary, err := Process(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, 0, summary.ExitCode())

	pkg, ok := summary.Packages["example.com/pkg"]
	require.True(t, ok)
	assert.False(t, pkg.HasPanic)
	assert.Equal(t, ActionPass, pkg.Summary.Action)

	test := pkg.GetTest("TestRecovered")
	require.NotNil(t, test)
	assert.Equal(t, ActionPass, test.Status())
	assert.Len(t, pkg.TestsByAction(ActionPass), 1)
	assert.Empty(t, pkg.TestsByAction(ActionFail))
}

func TestUnhandledPanicTestFail(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		`{"Time":"2025-09-09T11:00:00.000000-07:00","Action":"run","Package":"example.com/pkg","Test":"TestCrash"}`,
		`{"Time":"2025-09-09T11:00:00.000100-07:00","Action":"output","Package":"example.com/pkg","Test":"TestCrash","Output":"=== RUN   TestCrash\n"}`,
		`{"Time":"2025-09-09T11:00:00.000200-07:00","Action":"output","Package":"example.com/pkg","Test":"TestCrash","Output":"panic: nil pointer dereference\n"}`,
		`{"Time":"2025-09-09T11:00:00.000300-07:00","Action":"output","Package":"example.com/pkg","Test":"TestCrash","Output":"goroutine 1 [running]:\n"}`,
		`{"Time":"2025-09-09T11:00:00.000400-07:00","Action":"fail","Package":"example.com/pkg","Test":"TestCrash","Elapsed":0.01}`,
		`{"Time":"2025-09-09T11:00:00.000500-07:00","Action":"output","Package":"example.com/pkg","Output":"FAIL\texample.com/pkg\t0.010s\n"}`,
		`{"Time":"2025-09-09T11:00:00.000600-07:00","Action":"fail","Package":"example.com/pkg","Elapsed":0.01}`,
	}, "\n")

	summary, err := Process(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, 1, summary.ExitCode())

	pkg, ok := summary.Packages["example.com/pkg"]
	require.True(t, ok)
	assert.True(t, pkg.HasPanic)
	assert.Equal(t, ActionFail, pkg.Summary.Action)
	assert.Equal(t, "TestCrash", pkg.Summary.Test)
	assert.NotEmpty(t, pkg.PanicEvents)

	test := pkg.GetTest("TestCrash")
	require.NotNil(t, test)
	assert.Equal(t, ActionFail, test.Status())
	assert.Len(t, pkg.TestsByAction(ActionFail), 1)
}

func TestMixedRecoveredAndUnhandledPanics(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		`{"Time":"2025-09-09T11:00:00.000000-07:00","Action":"run","Package":"example.com/pkg","Test":"TestRecovered"}`,
		`{"Time":"2025-09-09T11:00:00.000100-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"=== RUN   TestRecovered\n"}`,
		`{"Time":"2025-09-09T11:00:00.000200-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"panic: expected panic [recovered]\n"}`,
		`{"Time":"2025-09-09T11:00:00.000300-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"--- PASS: TestRecovered (0.00s)\n"}`,
		`{"Time":"2025-09-09T11:00:00.000400-07:00","Action":"pass","Package":"example.com/pkg","Test":"TestRecovered","Elapsed":0}`,
		`{"Time":"2025-09-09T11:00:00.000500-07:00","Action":"run","Package":"example.com/pkg","Test":"TestFatalPanic"}`,
		`{"Time":"2025-09-09T11:00:00.000600-07:00","Action":"output","Package":"example.com/pkg","Test":"TestFatalPanic","Output":"=== RUN   TestFatalPanic\n"}`,
		`{"Time":"2025-09-09T11:00:00.000700-07:00","Action":"output","Package":"example.com/pkg","Test":"TestFatalPanic","Output":"panic: unhandled fatal panic\n"}`,
		`{"Time":"2025-09-09T11:00:00.000800-07:00","Action":"fail","Package":"example.com/pkg","Test":"TestFatalPanic","Elapsed":0.01}`,
		`{"Time":"2025-09-09T11:00:00.000900-07:00","Action":"output","Package":"example.com/pkg","Output":"FAIL\texample.com/pkg\t0.020s\n"}`,
		`{"Time":"2025-09-09T11:00:00.001000-07:00","Action":"fail","Package":"example.com/pkg","Elapsed":0.02}`,
	}, "\n")

	summary, err := Process(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, 1, summary.ExitCode())

	pkg, ok := summary.Packages["example.com/pkg"]
	require.True(t, ok)
	assert.True(t, pkg.HasPanic)
	assert.Equal(t, ActionFail, pkg.Summary.Action)
	assert.Equal(t, "TestFatalPanic", pkg.Summary.Test)

	testRecovered := pkg.GetTest("TestRecovered")
	require.NotNil(t, testRecovered)
	assert.Equal(t, ActionPass, testRecovered.Status())

	testFatal := pkg.GetTest("TestFatalPanic")
	require.NotNil(t, testFatal)
	assert.Equal(t, ActionFail, testFatal.Status())

	assert.Len(t, pkg.TestsByAction(ActionPass), 1)
	assert.Len(t, pkg.TestsByAction(ActionFail), 1)
}

func TestMixedRecoveredPanicAndNormalFail(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		`{"Time":"2025-09-09T11:00:00.000000-07:00","Action":"run","Package":"example.com/pkg","Test":"TestRecovered"}`,
		`{"Time":"2025-09-09T11:00:00.000100-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"=== RUN   TestRecovered\n"}`,
		`{"Time":"2025-09-09T11:00:00.000200-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"panic: handled panic [recovered]\n"}`,
		`{"Time":"2025-09-09T11:00:00.000300-07:00","Action":"output","Package":"example.com/pkg","Test":"TestRecovered","Output":"--- PASS: TestRecovered (0.00s)\n"}`,
		`{"Time":"2025-09-09T11:00:00.000400-07:00","Action":"pass","Package":"example.com/pkg","Test":"TestRecovered","Elapsed":0}`,
		`{"Time":"2025-09-09T11:00:00.000500-07:00","Action":"run","Package":"example.com/pkg","Test":"TestNormalFail"}`,
		`{"Time":"2025-09-09T11:00:00.000600-07:00","Action":"output","Package":"example.com/pkg","Test":"TestNormalFail","Output":"=== RUN   TestNormalFail\n"}`,
		`{"Time":"2025-09-09T11:00:00.000700-07:00","Action":"output","Package":"example.com/pkg","Test":"TestNormalFail","Output":"    fail_test.go:10: assertion failed\n"}`,
		`{"Time":"2025-09-09T11:00:00.000800-07:00","Action":"fail","Package":"example.com/pkg","Test":"TestNormalFail","Elapsed":0.01}`,
		`{"Time":"2025-09-09T11:00:00.000900-07:00","Action":"output","Package":"example.com/pkg","Output":"FAIL\texample.com/pkg\t0.020s\n"}`,
		`{"Time":"2025-09-09T11:00:00.001000-07:00","Action":"fail","Package":"example.com/pkg","Elapsed":0.02}`,
	}, "\n")

	summary, err := Process(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, 1, summary.ExitCode())

	pkg, ok := summary.Packages["example.com/pkg"]
	require.True(t, ok)
	assert.False(t, pkg.HasPanic)
	assert.Equal(t, ActionFail, pkg.Summary.Action)

	testRecovered := pkg.GetTest("TestRecovered")
	require.NotNil(t, testRecovered)
	assert.Equal(t, ActionPass, testRecovered.Status())

	testNormal := pkg.GetTest("TestNormalFail")
	require.NotNil(t, testNormal)
	assert.Equal(t, ActionFail, testNormal.Status())

	assert.Len(t, pkg.TestsByAction(ActionPass), 1)
	assert.Len(t, pkg.TestsByAction(ActionFail), 1)
}

func TestPackageLevelPanicUnhandled(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		`{"Time":"2025-09-09T11:00:00.000000-07:00","Action":"output","Package":"example.com/pkg","Output":"panic: init failure\n"}`,
		`{"Time":"2025-09-09T11:00:00.000100-07:00","Action":"output","Package":"example.com/pkg","Output":"goroutine 1 [running]:\n"}`,
		`{"Time":"2025-09-09T11:00:00.000200-07:00","Action":"output","Package":"example.com/pkg","Output":"FAIL\texample.com/pkg\t0.005s\n"}`,
		`{"Time":"2025-09-09T11:00:00.000300-07:00","Action":"fail","Package":"example.com/pkg","Elapsed":0.005}`,
	}, "\n")

	summary, err := Process(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, 1, summary.ExitCode())

	pkg, ok := summary.Packages["example.com/pkg"]
	require.True(t, ok)
	assert.True(t, pkg.HasPanic)
	assert.Equal(t, ActionFail, pkg.Summary.Action)
	assert.NotEmpty(t, pkg.PanicEvents)
}
