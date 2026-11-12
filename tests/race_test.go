package parsetest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/internal/app"
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
			inputFile := filepath.Join(base, tc.fileName+".jsonl")
			f, err := os.Open(inputFile)
			require.NoError(t, err)
			defer f.Close()

			summary, err := parse.Process(f)
			require.NoError(t, err)

			if summary.ExitCode() == 0 {
				t.Fatalf("expecting non-zero exit code")
			}
			for name, pkg := range summary.Packages {
				wantTestName, ok := tc.expected[name]
				if !ok {
					t.Fatalf("failed to find package: %q", name)
				}
				assert.Len(t, pkg.DataRaceTests, len(wantTestName))
				if len(pkg.DataRaceTests) > 0 {
					for i := range pkg.DataRaceTests {
						assert.Equal(t, pkg.DataRaceTests[i], wantTestName[i])
						test := pkg.GetTest(pkg.DataRaceTests[i])
						require.NotNil(t, test)
						assert.True(t, test.HasDataRace())
						assert.Equal(t, parse.ActionFail, test.Status())
					}
				}
			}
		})
	}
}

func TestRaceFailedTestsOutput(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "race")

	t.Run("test_level_race_captured_in_failed_tests", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		inputFile := filepath.Join(base, "test_01.jsonl")
		options := app.Options{
			FileName: inputFile,
			Output:   buf,
			Sorter:   parse.SortByPackageName,
		}
		gotExitCode, err := app.Run(options)
		require.NoError(t, err)
		assert.Equal(t, 1, gotExitCode)
		out := buf.String()
		assert.Contains(t, out, "TestA")
		assert.Contains(t, out, "WARNING: DATA RACE")
	})

	t.Run("package_level_race_without_failed_tests", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		inputFile := filepath.Join(base, "test_03.jsonl")
		options := app.Options{
			FileName: inputFile,
			Output:   buf,
			Sorter:   parse.SortByPackageName,
		}
		gotExitCode, err := app.Run(options)
		require.NoError(t, err)
		assert.Equal(t, 1, gotExitCode)
		out := buf.String()
		assert.Contains(t, out, "DATA RACE")
		assert.Contains(t, out, "debug/tparse-24")
	})

	t.Run("package_level_race_markdown", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		inputFile := filepath.Join(base, "test_03.jsonl")
		options := app.Options{
			FileName: inputFile,
			Output:   buf,
			Format:   app.OutputFormatMarkdown,
			Sorter:   parse.SortByPackageName,
		}
		gotExitCode, err := app.Run(options)
		require.NoError(t, err)
		assert.Equal(t, 1, gotExitCode)
		out := buf.String()
		assert.Contains(t, out, "DATA RACE")
		assert.Contains(t, out, "debug/tparse-24")
	})

	t.Run("package_level_race_no_tests", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		inputFile := filepath.Join(base, "test_06.jsonl")
		options := app.Options{
			FileName: inputFile,
			Output:   buf,
			Sorter:   parse.SortByPackageName,
		}
		gotExitCode, err := app.Run(options)
		require.NoError(t, err)
		assert.Equal(t, 1, gotExitCode)
		out := buf.String()
		assert.Contains(t, out, "DATA RACE")
		assert.Contains(t, out, "github.com/mfridman/debug-go/testing")
	})
}
