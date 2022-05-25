package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestPackageCache(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "cached")

	// Where bool indicates whether the package is expected to be marked as cached.
	type expected map[string]bool

	// Note, remember to clean the cache before running:
	// go clean -testcache
	tt := []struct {
		fileName string
		expected
	}{
		{
			// go test strings fmt -json
			// go test strings fmt time mime -json
			"test_01",
			expected{
				"strings": true,
				"fmt":     true,
				"time":    false,
				"mime":    false,
			},
		},
		{
			// go test log mime sort strings -json
			// go test bufio bytes crypto fmt log mime sort strings time -json
			"test_02",
			expected{
				"bufio":   false,
				"bytes":   false,
				"crypto":  false,
				"fmt":     false,
				"log":     true,
				"mime":    true,
				"sort":    true,
				"strings": true,
				"time":    false,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			intputFile := filepath.Join(base, tc.fileName+".json")
			f, err := os.Open(intputFile)
			check.NoError(t, err)

			summary, err := parse.Process(f)
			check.NoError(t, err)
			check.Number(t, len(summary.Packages), len(tc.expected))

			for name, pkg := range summary.Packages {
				t.Run(name, func(t *testing.T) {
					wantCached, ok := tc.expected[name]
					if !ok {
						t.Fatalf("got unexpected package name: %q", name)
					}
					if pkg.Cached != wantCached {
						t.Fatalf("got %t, want package %q to have cached field marked %t", pkg.Cached, name, wantCached)
					}
				})
			}
		})
	}
}
