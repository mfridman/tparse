package parsetest

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/parse"
)

func TestPackageStartTime(t *testing.T) {
	t.Parallel()

	// This test depends on go120_start_action.jsonl, which contains test output from go1.20

	expected := map[string]string{
		"github.com/pressly/goose/v4":                                          "2023-05-28T18:36:01.280967-04:00",
		"github.com/pressly/goose/v4/internal/check":                           "2023-05-28T18:36:01.281088-04:00",
		"github.com/pressly/goose/v4/internal/cli":                             "2023-05-28T18:36:01.281147-04:00",
		"github.com/pressly/goose/v4/internal/dialectadapter":                  "2023-05-28T18:36:01.281218-04:00",
		"github.com/pressly/goose/v4/internal/dialectadapter/dialectquery":     "2023-05-28T18:36:01.281253-04:00",
		"github.com/pressly/goose/v4/internal/migration":                       "2023-05-28T18:36:01.281269-04:00",
		"github.com/pressly/goose/v4/internal/migrationstats":                  "2023-05-28T18:36:01.281381-04:00",
		"github.com/pressly/goose/v4/internal/migrationstats/migrationstatsos": "2023-05-28T18:36:01.281426-04:00",
		"github.com/pressly/goose/v4/internal/normalizedsn":                    "2023-05-28T18:36:01.281465-04:00",
		"github.com/pressly/goose/v4/internal/sqlparser":                       "2023-05-28T18:36:01.446915-04:00",
		"github.com/pressly/goose/v4/internal/testdb":                          "2023-05-28T18:36:01.446973-04:00",
	}

	fileName := "./testdata/go120_start_action.jsonl"
	f, err := os.Open(fileName)
	require.NoError(t, err)
	defer f.Close()

	summary, err := parse.Process(f)
	require.NoError(t, err)
	assert.Equal(t, len(summary.Packages), len(expected))

	for _, p := range summary.Packages {
		if p.StartTime.IsZero() {
			t.Fatalf("package %q cannot contain zero start time", p.Summary.Package)
		}
		unparsed, ok := expected[p.Summary.Package]
		if !ok {
			t.Fatalf("package %q not found in expected map", p.Summary.Package)
		}
		want, err := time.Parse(time.RFC3339, unparsed)
		require.NoError(t, err)
		if !p.StartTime.Equal(want) {
			t.Fatalf("package %q start time got %q want %q", p.Summary.Package, p.StartTime, want)
		}
	}
}
