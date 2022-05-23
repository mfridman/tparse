package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestPrescan(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "prescan")

	tt := []struct {
		fileName string
		desc     string
		err      error
	}{
		{"test_01.txt", "want <nil> err", nil},
		{"test_02.txt", "want failure after reading >50 lines of non-parseable events", parse.ErrNotParseable},
		// logic: unparseable event(s), good event(s), at least one event = fail.
		// Once we get a good event, we expect only good events to follow until EOF.
		{"test_03.txt", "want failure when stream contains a bad event(s) -> good event(s) -> bad event", parse.ErrNotParseable},
		{"test_04.txt", "want failure reading <50 lines of non-parseable events", parse.ErrNotParseable},
	}

	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			intputFile := filepath.Join(base, tc.fileName)
			f, err := os.Open(intputFile)
			check.NoError(t, err)

			_, err = parse.Process(f)
			check.IsError(t, err, tc.err)
		})

	}
}
