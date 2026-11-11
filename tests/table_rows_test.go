package parsetest

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/parse"
)

// TestTestsTableRows checks that the tests table has no empty rows at the
// start or the end, and no two empty rows in a row. Packages that add no rows
// to the table must not add a blank row. Examples are packages without test
// files, or packages with only skipped tests when skipped tests are hidden.
//
// The test data comes from a module with these packages:
//
//	a_nofiles   code, no test files
//	b_tested    2 passing tests
//	c_nofiles   code, no test files
//	d_onlyskip  1 skipped test
//	e_tested    2 passing tests
//	f_nofiles   code, no test files
//
// With -cover, Go reports packages without test files as 0.0% coverage, and
// not as [no test files], so they are not removed from the table early.
func TestTestsTableRows(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "table-rows")

	tt := []struct {
		fileName   string
		pass, skip bool
		// wantGroups is the number of packages with at least one row.
		wantGroups int
	}{
		// go test -count=1 -json -cover ./...
		{fileName: "test_01", pass: true, skip: true, wantGroups: 3},
		{fileName: "test_01", pass: true, wantGroups: 2},
		{fileName: "test_01", skip: true, wantGroups: 1},
		// go test -count=1 -json ./...
		{fileName: "test_02", pass: true, skip: true, wantGroups: 3},
		{fileName: "test_02", pass: true, wantGroups: 2},
		{fileName: "test_02", skip: true, wantGroups: 1},
		// go test -count=1 -json -cover ./b_tested ./c_nofiles
		{fileName: "test_03", pass: true, skip: true, wantGroups: 1},
		{fileName: "test_03", skip: true, wantGroups: 0},
	}
	for _, tc := range tt {
		name := tc.fileName
		if tc.pass {
			name += "/pass"
		}
		if tc.skip {
			name += "/skip"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			_, err := app.Run(app.Options{
				FileName:     filepath.Join(base, tc.fileName+".jsonl"),
				Output:       &buf,
				DisableColor: true,
				Sorter:       parse.SortByPackageName,
				TestTableOptions: app.TestTableOptions{
					Pass: tc.pass,
					Skip: tc.skip,
				},
			})
			require.NoError(t, err)

			rows := testsTableRows(t, buf.String())
			if tc.wantGroups == 0 {
				require.Empty(t, rows)
				return
			}
			table := strings.Join(rows, "\n")
			require.NotEmpty(t, rows)
			require.False(t, isEmptyRow(rows[0]), "first row is empty:\n%s", table)
			require.False(t, isEmptyRow(rows[len(rows)-1]), "last row is empty:\n%s", table)
			var blank int
			for i, row := range rows {
				if isEmptyRow(row) {
					blank++
					require.False(t, isEmptyRow(rows[i-1]), "two empty rows in a row:\n%s", table)
				}
			}
			require.Equal(t, tc.wantGroups-1, blank, "want one blank row between packages:\n%s", table)
		})
	}
}

// testsTableRows returns the body rows of the tests table. The tests table is
// the first table in the output. It returns nil if there is no tests table.
func testsTableRows(t *testing.T, output string) []string {
	t.Helper()

	lines := strings.Split(output, "\n")
	// Line 0 is the top border, line 1 is the header, and line 2 is the
	// border below the header. The summary table header has a Cover column.
	if len(lines) < 3 || !strings.Contains(lines[1], "Test") || strings.Contains(lines[1], "Cover") {
		return nil
	}
	var rows []string
	for _, line := range lines[3:] {
		if strings.HasPrefix(line, "╰") {
			return rows
		}
		rows = append(rows, line)
	}
	require.FailNow(t, "tests table has no bottom border", output)
	return nil
}

func isEmptyRow(row string) bool {
	return strings.Trim(row, "│ ") == ""
}
