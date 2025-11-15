package app

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/mfridman/tparse/internal/utils"
	"github.com/mfridman/tparse/parse"
)

var (
	versionMajorRe = regexp.MustCompile(`(?m)v[0-9]+`)
)

type TestTableOptions struct {
	// Display passed or skipped tests. If both are true this is equivalent to all.
	Pass, Skip bool
	// For narrow screens, trim long test identifiers vertically. Example:
	// TestNoVersioning/seed-up-down-to-zero
	//
	// TestNoVersioning
	//  /seed-up-down-to-zero
	Trim bool

	// TrimPath is the path prefix to trim from the package name.
	TrimPath string

	// Display up to N slow tests for each package, tests are sorted by
	// calculated the elapsed time for the given test.
	Slow int
}

type packageTests struct {
	skippedCount int
	skipped      []*parse.Test
	passedCount  int
	passed       []*parse.Test
	failed       []*parse.Test
	failedCount  int
}

func (c *consoleWriter) testsTable(packages []*parse.Package, option TestTableOptions) {
	// Print passed tests, sorted by elapsed DESC. Grouped by alphabetically sorted packages.
	tbl := newTable(c.format, func(style lipgloss.Style, row, col int) lipgloss.Style {
		switch row {
		case table.HeaderRow:
		default:
			if col == 2 || col == 3 {
				// Test name and package name
				style = style.Align(lipgloss.Left)
			}
		}
		return style
	})
	header := testRow{
		status:      "Status",
		elapsed:     "Elapsed",
		testName:    "Test",
		packageName: "Package",
	}
	tbl.Headers(header.toRow()...)
	data := table.NewStringData()

	names := make([]string, 0, len(packages))
	for _, pkg := range packages {
		names = append(names, pkg.Summary.Package)
	}
	packagePrefix := utils.FindLongestCommonPrefix(names)

	for i, pkg := range packages {
		// Discard packages where we cannot generate a sensible test summary.
		if pkg.NoTestFiles || pkg.NoTests || pkg.HasPanic {
			continue
		}
		pkgTests := getTestsFromPackages(pkg, option)
		all := make([]*parse.Test, 0, len(pkgTests.passed)+len(pkgTests.skipped)+len(pkgTests.failed))
		all = append(all, pkgTests.passed...)
		all = append(all, pkgTests.skipped...)
		all = append(all, pkgTests.failed...)

		for _, t := range all {
			// TODO(mf): why are we sorting this?
			t.SortEvents()

			testName := shortenTestName(t.Name, option.Trim, 32)

			status := strings.ToUpper(t.Status().String())
			switch t.Status() {
			case parse.ActionPass:
				status = c.green(status)
			case parse.ActionSkip:
				status = c.yellow(status)
			case parse.ActionFail:
				status = c.red(status)
			}

			packageName := shortenPackageName(t.Package, packagePrefix, 16, option.Trim, option.TrimPath)

			row := testRow{
				status:      status,
				elapsed:     strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName:    testName,
				packageName: packageName,
			}
			data.Append(row.toRow())
		}
		if i != (len(packages) - 1) {
			// Add a blank row between packages.
			data.Append(testRow{}.toRow())
		}
	}

	if data.Rows() > 0 {
		fmt.Fprintln(c.w, tbl.Data(data).Render())
	}
}

func (c *consoleWriter) testsTableMarkdown(packages []*parse.Package, option TestTableOptions) {
	for _, pkg := range packages {
		// Print passed tests, sorted by elapsed DESC. Grouped by alphabetically sorted packages.
		tbl := newTable(c.format, func(style lipgloss.Style, row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
			default:
				if col == 2 {
					// Test name
					style = style.Align(lipgloss.Left)
				}
			}
			return style
		})
		header := []string{
			"Status",
			"Elapsed",
			"Test",
		}
		tbl.Headers(header...)
		data := table.NewStringData()

		// Discard packages where we cannot generate a sensible test summary.
		if pkg.NoTestFiles || pkg.NoTests || pkg.HasPanic {
			continue
		}
		pkgTests := getTestsFromPackages(pkg, option)
		all := make([]*parse.Test, 0, len(pkgTests.passed)+len(pkgTests.skipped)+len(pkgTests.failed))
		all = append(all, pkgTests.passed...)
		all = append(all, pkgTests.skipped...)
		all = append(all, pkgTests.failed...)

		for _, t := range all {
			// TODO(mf): why are we sorting this?
			t.SortEvents()

			testName := shortenTestName(t.Name, option.Trim, 32)

			status := strings.ToUpper(t.Status().String())
			switch t.Status() {
			case parse.ActionPass:
				status = c.green(status)
			case parse.ActionSkip:
				status = c.yellow(status)
			case parse.ActionFail:
				status = c.red(status)
			}
			data.Append([]string{
				status,
				strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName,
			})
		}
		if data.Rows() > 0 {
			fmt.Fprintf(c.w, "## 📦 Package **`%s`**\n", pkg.Summary.Package)
			fmt.Fprintln(c.w)

			msg := fmt.Sprintf("Tests: ✓ %d passed | %d skipped | %d failed\n",
				pkgTests.passedCount,
				pkgTests.skippedCount,
				pkgTests.failedCount,
			)
			if option.Slow > 0 && option.Slow < pkgTests.passedCount {
				msg += fmt.Sprintf("↓ Slowest %d passed tests shown (of %d)\n",
					option.Slow,
					pkgTests.passedCount,
				)
			}
			fmt.Fprint(c.w, msg)

			fmt.Fprintln(c.w)
			fmt.Fprintln(c.w, "<details>")
			fmt.Fprintln(c.w)
			fmt.Fprintln(c.w, "<summary>Click for test summary</summary>")
			fmt.Fprintln(c.w)
			fmt.Fprintln(c.w, tbl.Data(data).Render())
			fmt.Fprintln(c.w, "</details>")
			fmt.Fprintln(c.w)
		}
		fmt.Fprintln(c.w)
	}
}

func getTestsFromPackages(pkg *parse.Package, option TestTableOptions) *packageTests {
	tests := &packageTests{}
	skipped := pkg.TestsByAction(parse.ActionSkip)
	tests.skippedCount = len(skipped)
	passed := pkg.TestsByAction(parse.ActionPass)
	tests.passedCount = len(passed)
	failed := pkg.TestsByAction(parse.ActionFail)
	tests.failedCount = len(failed)
	if option.Skip {
		tests.skipped = append(tests.skipped, skipped...)
	}
	if option.Pass {
		tests.passed = append(tests.passed, passed...)
		// Order passed tests within a package by elapsed time DESC (longest on top).
		sort.Slice(tests.passed, func(i, j int) bool {
			return tests.passed[i].Elapsed() > tests.passed[j].Elapsed()
		})
		// Optional, display only the slowest N tests by elapsed time.
		if option.Slow > 0 && len(tests.passed) > option.Slow {
			tests.passed = tests.passed[:option.Slow]
		}
	}
	tests.failed = append(tests.failed, failed...)
	return tests
}

func shortenTestName(s string, trim bool, maxLength int) string {
	var testName strings.Builder
	testName.WriteString(s)
	if trim && testName.Len() > maxLength && strings.Count(testName.String(), "/") > 0 {
		testName.Reset()
		ss := strings.Split(s, "/")
		testName.WriteString(ss[0] + "\n")
		for i, s := range ss[1:] {
			testName.WriteString(" /")
			for len(s) > maxLength {
				testName.WriteString(s[:maxLength-2] + " …\n  ")
				s = s[maxLength-2:]
			}
			testName.WriteString(s)
			if i != len(ss[1:])-1 {
				testName.WriteString("\n")
			}
		}
	}
	return testName.String()
}

type testRow struct {
	status      string
	elapsed     string
	testName    string
	packageName string
}

func (r testRow) toRow() []string {
	return []string{
		r.status,
		r.elapsed,
		r.testName,
		r.packageName,
	}
}
