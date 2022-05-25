package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func Test(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "cover")

	// expected package name and corresponding cover %
	// Note, these numbers will vary between on go versions.
	type expected map[string]float64

	tt := []struct {
		fileName string
		expected
	}{
		{
			// go test -count=1 bytes log sort -json -cover
			"test_01",
			expected{"log": 68.0, "bytes": 86.7, "sort": 60.8},
		},
		{
			// go test -count=1 bufio bytes crypto fmt log mime net sort strings time -json -cover
			"test_02",
			expected{
				"bufio":   93.3,
				"bytes":   95.6,
				"crypto":  5.9,
				"fmt":     95.2,
				"log":     68.0,
				"mime":    93.8,
				"net":     81.2,
				"sort":    60.8,
				"strings": 98.1,
				"time":    91.8,
			},
		},
		{
			// This is run without the -cover flag. Expecting 0.0 for all packages.
			// go test -count=1 crypto fmt log strings -json -cover
			"test_03",
			expected{
				"crypto":  0.0,
				"fmt":     0.0,
				"log":     0.0,
				"strings": 0.0,
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
					wantCover, ok := tc.expected[name]
					if !ok {
						t.Fatalf("got unexpected package name: %q", name)
					}
					if pkg.Coverage != wantCover {
						t.Fatalf("got cover: %v, want package %q cover: %v", pkg.Coverage, name, wantCover)
					}
					var f float64
					if pkg.Coverage > f && !pkg.Cover {
						t.Fatalf("got %v, want package %q to have cover field marked as true when coverage %v>%v",
							pkg.Cover,
							name,
							pkg.Coverage,
							f,
						)
					}
				})
			}
		})
	}
}
