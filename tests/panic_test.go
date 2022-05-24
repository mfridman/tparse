package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
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
			"test_01.json", expected{
				"github.com/pressly/goose/v3/tests/e2e": true,
			},
		},
		{
			"test_02.json", expected{
				"github.com/mfridman/tparse/parse": true,
			},
		},
		{
			"test_03.json", expected{
				"github.com/mfridman/tparse/tests": true,
			},
		},
		{
			"test_04.json", expected{
				"github.com/mfridman/tparse/tests":  true,
				"github.com/mfridman/tparse/ignore": false,
				"github.com/mfridman/tparse/parse":  false,
				"github.com/mfridman/tparse":        false,
			},
		},
		{
			"test_05.json", expected{
				"github.com/mfridman/tparse/tests":  true,
				"github.com/mfridman/tparse/parse":  false,
				"github.com/mfridman/tparse":        false,
				"github.com/mfridman/tparse/ignore": false,
			},
		},
		{
			"test_06.json", expected{
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
			check.NoError(t, err)

			summary, err := parse.Process(f)
			check.NoError(t, err)
			check.Number(t, summary.ExitCode(), 1)

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
