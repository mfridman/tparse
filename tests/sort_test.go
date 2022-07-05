package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestSortName(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "cover")

	// expected package name and corresponding cover %
	// Note, these numbers will vary between on go versions.
	type expected []string

	tt := []struct {
		fileName string
		expected
	}{
		{
			// go test -count=1 bytes log sort -json -cover | tparse -sort name
			"test_01",
			expected{
				"bytes",
				"log",
				"sort",
			},
		},
		{
			// go test -count=1 bufio bytes crypto fmt log mime net sort strings time -json -cover | tparse -sort name
			"test_02",
			expected{
				"bufio",
				"bytes",
				"crypto",
				"fmt",
				"log",
				"mime",
				"net",
				"sort",
				"strings",
				"time",
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
			packages := summary.GetSortedPackages(parse.SortByPackageName)

			for i, pkg := range packages {
				t.Run(pkg.Summary.Package, func(t *testing.T) {
					wantName := tc.expected[i]
					if pkg.Summary.Package != wantName {
						t.Fatalf("got name: %s, want name: %s", pkg.Summary.Package, wantName)
					}
				})
			}
		})
	}
}

func TestSortCoverage(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "cover")

	// expected package name and corresponding cover %
	// Note, these numbers will vary between on go versions.
	type expected []float64

	tt := []struct {
		fileName string
		expected
	}{
		{
			// go test -count=1 bytes log sort -json -cover | tparse -sort cover
			"test_01",
			expected{
				86.7, // "bytes"
				68.0, // "log"
				60.8, // "sort"
			},
		},
		{
			// go test -count=1 bufio bytes crypto fmt log mime net sort strings time -json -cover | tparse -sort cover
			"test_02",
			expected{
				98.1, // "strings"
				95.6, // "bytes"
				95.2, // "fmt"
				93.8, // "mime"
				93.3, // "bufio"
				91.8, // "time"
				81.2, // "net"
				68.0, // "log"
				60.8, // "sort"
				5.9,  // "crypto"
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
			packages := summary.GetSortedPackages(parse.SortByCoverage)

			for i, pkg := range packages {
				t.Run(pkg.Summary.Package, func(t *testing.T) {
					wantCover := tc.expected[i]
					if pkg.Coverage != wantCover {
						t.Fatalf("got cover: %v(%s), want cover: %v", pkg.Coverage, pkg.Summary.Package, wantCover)
					}
				})
			}
		})
	}
}

func TestSortElapsed(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "cached")

	// expected package name and corresponding cover %
	// Note, these numbers will vary between on go versions.
	type expected []float64

	tt := []struct {
		fileName string
		expected
	}{
		{
			// go test -count=1 fmt mime strings time -json | tparse -sort elapsed
			"test_01",
			expected{
				7.168, // "time"
				0.020, // "mime"
				0.007, // "strings"
				0.003, // "fmt"
			},
		},
		{
			// go test -count=1 bufio bytes crypto fmt log mime sort strings time -json | tparse -sort elapsed
			"test_02",
			expected{
				7.641, // "time",
				1.176, // "bytes",
				0.220, // "fmt",
				0.134, // "bufio",
				0.070, // "crypto",
				0.002, // "strings",
				0.001, // "mime",
				0.001, // "sort",
				0.000, // "log",
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
			packages := summary.GetSortedPackages(parse.SortByElapsed)

			for i, pkg := range packages {
				t.Run(pkg.Summary.Package, func(t *testing.T) {
					wantElapsed := tc.expected[i]
					if pkg.Summary.Elapsed != wantElapsed {
						t.Fatalf("got elapsed: %v (%s), want elapsed: %v", pkg.Summary.Elapsed, pkg.Summary.Package, wantElapsed)
					}
				})
			}
		})
	}
}
