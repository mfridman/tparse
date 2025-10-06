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

func TestFailedTestsTable(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "failed")

	tt := []struct {
		fileName string
		exitCode int
	}{
		{"test_01", 1},
		{"test_02", 1},
		{"test_03", 1},
		{"test_04", 1},
	}

	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			inputFile := filepath.Join(base, tc.fileName+".jsonl")
			options := app.Options{
				FileName: inputFile,
				Output:   buf,
				Sorter:   parse.SortByPackageName,
				TestTableOptions: app.TestTableOptions{
					Pass: true, // Enable test table output
					Skip: true, // Also show skipped tests
				},
			}
			gotExitCode, err := app.Run(options)
			require.NoError(t, err)
			assert.Equal(t, tc.exitCode, gotExitCode)

			goldenFile := filepath.Join(base, tc.fileName+".golden")
			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatal(err)
			}
			checkGolden(t, inputFile, goldenFile, buf.Bytes(), want)
		})
	}
}
