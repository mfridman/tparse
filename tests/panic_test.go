package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/internal/parse"
)

func TestPanic(t *testing.T) {
	t.Parallel()

	// Key is the package name, bool reports whether the packages is expected to be marked as panic.
	type expected map[string]bool

	base := filepath.Join("testdata", "panic")

	tt := []struct {
		fileName string
		expected
	}{
		{
			"test_01.jsonl", expected{
				"github.com/pressly/goose/v3/tests/e2e": true,
			},
		},
		{
			"test_02.jsonl", expected{
				"github.com/mfridman/tparse/parse": true,
			},
		},
		{
			"test_03.jsonl", expected{
				"github.com/mfridman/tparse/tests": true,
			},
		},
		{
			"test_04.jsonl", expected{
				"github.com/mfridman/tparse/tests":  true,
				"github.com/mfridman/tparse/ignore": false,
				"github.com/mfridman/tparse/parse":  false,
				"github.com/mfridman/tparse":        false,
			},
		},
		{
			"test_05.jsonl", expected{
				"github.com/mfridman/tparse/tests":  true,
				"github.com/mfridman/tparse/parse":  false,
				"github.com/mfridman/tparse":        false,
				"github.com/mfridman/tparse/ignore": false,
			},
		},
		{
			"test_06.jsonl", expected{
				"github.com/mfridman/tparse/tests":  false,
				"github.com/mfridman/tparse/parse":  true,
				"github.com/mfridman/tparse":        false,
				"github.com/mfridman/tparse/ignore": false,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			f, err := os.Open(filepath.Join(base, tc.fileName))
			require.NoError(t, err)

			summary, err := parse.Process(f)
			require.NoError(t, err)
			assert.Equal(t, 1, summary.ExitCode())

			for name, pkg := range summary.Packages {
				want, ok := tc.expected[name]
				if !ok {
					t.Log("currently registered packages:")
					for k := range tc.expected {
						t.Log(k)
					}
					t.Fatalf("got unmapped package name %q. Check input file and record all unique package names in registry", name)
				}
				if pkg.HasPanic != want {
					t.Log("package: ", name)
					t.Logf("summary: %+v", pkg.Summary)
					t.Fatal("got no panic, expecting package to be marked as has panic")
				}
			}
		})
	}
}
