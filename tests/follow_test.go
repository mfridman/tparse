package parsetest

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
