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

// TestBuildFailure checks build and setup failures in the go test -json format
// of go1.24 and later. In this format, the go command reports these failures as
// JSON events, and not as raw text lines.
func TestBuildFailure(t *testing.T) {
	t.Parallel()

	inputFile := filepath.Join("testdata", "build-failure", "test_01.jsonl")

	t.Run("summary", func(t *testing.T) {
		t.Parallel()
		inputReader, err := os.Open(inputFile)
		require.NoError(t, err)
		defer inputReader.Close()

		summary, err := parse.Process(inputReader)
		require.NoError(t, err)
		assert.Equal(t, 2, summary.ExitCode())

		wantReasons := map[string]string{
			"example.com/demo/broken": "build failed",
			"example.com/demo/setup":  "setup failed",
		}
		require.Len(t, summary.Packages, 4)
		for name, pkg := range summary.Packages {
			wantReason, wantFailure := wantReasons[name]
			assert.Equal(t, wantFailure, pkg.HasFailedBuildOrSetup, name)
			if wantFailure {
				assert.Equal(t, parse.ActionFail, pkg.Summary.Action, name)
				assert.Equal(t, wantReason, pkg.Summary.Output, name)
			}
		}
	})

	t.Run("app", func(t *testing.T) {
		t.Parallel()
		var output bytes.Buffer
		gotExitCode, err := app.Run(app.Options{
			FileName:     inputFile,
			Output:       &output,
			DisableColor: true,
			Sorter:       parse.SortByPackageName,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, gotExitCode)
		assert.Contains(t, output.String(), "[build failed]")
		assert.Contains(t, output.String(), "[setup failed]")
	})
}

// TestExitCodeBuildFailureFirst checks that a build or setup failure sets exit
// code 2, also when the package map gives another failed package first.
func TestExitCodeBuildFailureFirst(t *testing.T) {
	t.Parallel()

	summary := &parse.GoTestSummary{
		Packages: map[string]*parse.Package{
			"example.com/failed": {
				Summary: &parse.Event{Action: parse.ActionFail},
			},
			"example.com/panic": {
				Summary:  &parse.Event{Action: parse.ActionFail},
				HasPanic: true,
			},
			"example.com/build": {
				Summary:               &parse.Event{Action: parse.ActionFail},
				HasFailedBuildOrSetup: true,
			},
		},
	}
	// Map iteration order is random, so run many times to cover each order.
	for range 100 {
		require.Equal(t, 2, summary.ExitCode())
	}
}
