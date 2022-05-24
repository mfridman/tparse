package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestRaceDetected(t *testing.T) {
	t.Parallel()

	// Key is the package name, and the value may be zero, one or more test name(s).
	// Not all data races may be associated with a test.
	type expected map[string][]string

	base := filepath.Join("testdata", "race")

	var tt = []struct {
		fileName string
		expected
	}{
		{
			"test_01", expected{"command-line-arguments": {"TestA"}},
		},
		{
			"test_02", expected{"github.com/mfridman/tparse/parse": {"TestB", "TestElapsed"}},
		},
		{
			"test_03", expected{"debug/tparse-24": {}},
		},
		// This is a race directly from Test only.
		{
			"test_04", expected{"github.com/mfridman/debug-go/testing": {"TestRace"}},
		},
		// This is a race directly from TestMain with other tests that have failed.
		{
			"test_05", expected{"github.com/mfridman/debug-go/testing": {}},
		},
		// This is a race directly from TestMain only.
		{
			"test_06", expected{"github.com/mfridman/debug-go/testing": {}},
		},
		// This is a race from a Test that calls into a package that has a race condition. (failed assertion)
		{
			"test_07", expected{"github.com/mfridman/debug-go/testing": {"TestRace"}},
		},
		// This is a race from a Test that calls into a package that has a race condition. (passed assertion)
		{
			"test_08", expected{"github.com/mfridman/debug-go/testing": {"TestRace"}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			intputFile := filepath.Join(base, tc.fileName+".json")
			f, err := os.Open(intputFile)
			check.NoError(t, err)
			defer f.Close()

			summary, err := parse.Process(f)
			check.NoError(t, err)

			if summary.ExitCode() == 0 {
				t.Fatalf("expecting non-zero exit code")
			}
			for name, pkg := range summary.Packages {
				wantTestName, ok := tc.expected[name]
				if !ok {
					t.Fatalf("failed to find package: %q", name)
				}
				check.Number(t, len(pkg.DataRaceTests), len(wantTestName))
				if len(pkg.DataRaceTests) > 0 {
					for i := range pkg.DataRaceTests {
						check.Equal(t, pkg.DataRaceTests[i], wantTestName[i])
					}
				}
			}
		})
	}
}
