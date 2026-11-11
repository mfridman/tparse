package parsetest

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/parse"
)

func TestDisableColor(t *testing.T) {
	t.Parallel()

	// The failed test files render every styled element: colored status
	// cells, bordered package headers, and highlighted failure lines.
	inputFile := filepath.Join("testdata", "failed", "test_01.jsonl")
	for _, format := range []app.OutputFormat{
		app.OutputFormatBasic,
		app.OutputFormatPlain,
		app.OutputFormatMarkdown,
	} {
		var colored, uncolored bytes.Buffer
		for _, disableColor := range []bool{false, true} {
			output := &colored
			if disableColor {
				output = &uncolored
			}
			_, err := app.Run(app.Options{
				FileName:     inputFile,
				Output:       output,
				DisableColor: disableColor,
				Format:       format,
				Sorter:       parse.SortByPackageName,
				TestTableOptions: app.TestTableOptions{
					Pass: true,
					Skip: true,
				},
			})
			require.NoError(t, err)
		}
		if format != app.OutputFormatMarkdown {
			assert.Contains(t, colored.String(), "\x1b[", "format %d: expected ANSI sequences", format)
		}
		assert.NotContains(t, uncolored.String(), "\x1b", "format %d: expected no ANSI sequences", format)
	}
}
