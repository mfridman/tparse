package parsetest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/parse"
)

// TestNoTestFiles checks that packages without test files are found with and
// without -cover. Since go1.22, go test -cover reports a package that has code
// but no test files as "coverage: 0.0% of statements", and not as
// "[no test files]".
//
// The test data comes from a module with these packages:
//
//	emptytests  test file without tests
//	nocode      no test files, and no code (only types and constants)
//	notests     no test files
//	tested      1 passing test, 100% coverage
//	zerocov     1 passing test, 0% coverage
func TestNoTestFiles(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "no-test-files")
	const prefix = "example.com/cov106/"

	tt := []struct {
		fileName string
		comment  string
	}{
		{"test_01", "go test -count=1 -json ./..."},
		{"test_02", "go test -count=1 -json -cover ./..."},
		{"test_03", "go test -count=1 -json -coverpkg=./... ./..."},
	}
	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			t.Parallel()

			inputFile := filepath.Join(base, tc.fileName+".jsonl")
			f, err := os.Open(inputFile)
			require.NoError(t, err)
			t.Cleanup(func() { _ = f.Close() })
			summary, err := parse.Process(f)
			require.NoError(t, err)

			got := make(map[string]bool)
			for name, pkg := range summary.Packages {
				got[strings.TrimPrefix(name, prefix)] = pkg.NoTestFiles
			}
			want := map[string]bool{
				"emptytests": false,
				"nocode":     true,
				"notests":    true,
				"tested":     false,
				"zerocov":    false,
			}
			require.Equal(t, want, got, tc.comment)

			// With -notests, the summary table must show both packages without
			// test files. Without -notests, it must show neither.
			for _, showNoTests := range []bool{true, false} {
				var buf bytes.Buffer
				_, err = app.Run(app.Options{
					FileName:     inputFile,
					Output:       &buf,
					DisableColor: true,
					ShowNoTests:  showNoTests,
					Sorter:       parse.SortByPackageName,
				})
				require.NoError(t, err)
				for _, name := range []string{"nocode", "notests"} {
					row := findRow(buf.String(), prefix+name)
					if !showNoTests {
						require.Empty(t, row, "%s: want no row for %s:\n%s", tc.comment, name, buf.String())
						continue
					}
					require.Contains(t, row, "NOTEST", "%s: want NOTEST row for %s:\n%s", tc.comment, name, buf.String())
				}
			}
		})
	}
}

// findRow returns the first line in output that contains the package name
// followed by a space or a column border, or an empty string.
func findRow(output, packageName string) string {
	for line := range strings.Lines(output) {
		if strings.Contains(line, packageName+" ") || strings.Contains(line, packageName+"│") {
			return line
		}
	}
	return ""
}
