package parsetest

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/internal/utils"
	"github.com/mfridman/tparse/parse"
)

func TestFollow(t *testing.T) {
	t.Parallel()

	t.Run("follow_verbose", func(t *testing.T) {
		base := filepath.Join("testdata", "follow-verbose")

		tt := []struct {
			fileName string
			err      error
			exitCode int
		}{
			// race detected
			{"test_01", nil, 1},
			{"test_02", nil, 0},
			{"test_03", nil, 0},
			{"test_04", nil, 0},
			{"test_05", parse.ErrNotParsable, 1},
			// build failure in one package
			{"test_06", nil, 2},
		}
		for _, tc := range tt {
			t.Run(tc.fileName, func(t *testing.T) {
				buf := bytes.NewBuffer(nil)
				inputFile := filepath.Join(base, tc.fileName+".jsonl")
				options := app.Options{
					FileName:            inputFile,
					FollowOutput:        true,
					FollowOutputWriter:  utils.WriteNopCloser{Writer: buf},
					FollowOutputVerbose: true,
					DisableTableOutput:  true,
				}
				gotExitCode, err := app.Run(options)
				if err != nil && !errors.Is(err, tc.err) {
					t.Fatal(err)
				}
				assert.Equal(t, gotExitCode, tc.exitCode)
				goldenFile := filepath.Join(base, tc.fileName+".golden")
				want, err := os.ReadFile(goldenFile)
				if err != nil {
					t.Fatal(err)
				}
				checkGolden(t, inputFile, goldenFile, buf.Bytes(), want)
			})
		}
	})

	t.Run("follow_no_verbose", func(t *testing.T) {
		base := filepath.Join("testdata", "follow")

		tt := []struct {
			fileName string
			err      error
			exitCode int
		}{
			{"test_01", nil, 0},
		}
		for _, tc := range tt {
			t.Run(tc.fileName, func(t *testing.T) {
				buf := bytes.NewBuffer(nil)
				inputFile := filepath.Join(base, tc.fileName+".jsonl")
				options := app.Options{
					FileName:           inputFile,
					FollowOutput:       true,
					FollowOutputWriter: utils.WriteNopCloser{Writer: buf},
					DisableTableOutput: true,
				}
				gotExitCode, err := app.Run(options)
				if err != nil && !errors.Is(err, tc.err) {
					t.Fatal(err)
				}
				assert.Equal(t, gotExitCode, tc.exitCode)
				goldenFile := filepath.Join(base, tc.fileName+".golden")
				want, err := os.ReadFile(goldenFile)
				if err != nil {
					t.Fatal(err)
				}
				checkGolden(t, inputFile, goldenFile, buf.Bytes(), want)
			})
		}
	})
}

// The failing parent test includes a child go test transcript in its error message.
func TestFollowErrorOutput(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		follow  bool
		verbose bool
	}{
		{"default", false, false},
		{"follow", true, false},
		{"follow-verbose", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var report, live bytes.Buffer
			code, err := app.Run(app.Options{
				FileName:            filepath.Join("testdata", "go127_output_type.jsonl"),
				Output:              &report,
				Sorter:              parse.SortByPackageName,
				DisableColor:        true,
				FollowOutput:        tc.follow,
				FollowOutputVerbose: tc.verbose,
				FollowOutputWriter:  utils.WriteNopCloser{Writer: &live},
			})
			require.NoError(t, err)
			require.Equal(t, 1, code)

			const diagnostic = "outer_test.go:8: expected --- FAIL: TestChild in child transcript:\n" +
				"        === RUN   TestChild\n        --- PASS: TestChild (0.00s)\n"
			assert.Contains(t, report.String(), diagnostic)
			if tc.follow {
				assert.Contains(t, live.String(), diagnostic)
			}
			const outerFrame = "=== RUN   TestChildExpectedFailure\n"
			assert.Equal(t, tc.verbose, bytes.Contains(live.Bytes(), []byte(outerFrame)))
		})
	}
}

func checkGolden(
	t *testing.T,
	inputFile, goldenFile string,
	got, want []byte,
) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Error("input does not match expected output; diff files in follow dir suffixed with .FAIL to debug")
		t.Logf("diff %v %v",
			"tests/"+goldenFile+".FAIL",
			"tests/"+inputFile+".FAIL",
		)
		if err := os.WriteFile(goldenFile+".FAIL", got, 0644); err != nil {
			t.Fatal(err)
		}
	}
}
