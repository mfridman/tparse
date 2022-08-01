package app

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	// Display up to N slow tests for each package, tests are sorted by
	// calculated the elapsed time for the given test.
	Slow int
}

type packageTests struct {
	skipped []*parse.Test
	passed  []*parse.Test
}

func (c *consoleWriter) testsTable(packages []*parse.Package, option TestTableOptions) {
	// Print passed tests, sorted by elapsed DESC. Grouped by alphabetically sorted packages.
	var tableString strings.Builder
	tbl := newTableWriter(&tableString, c.format)

	header := testRow{
		status:      "Status",
		elapsed:     "Elapsed",
		testName:    "Test",
		packageName: "Package",
	}
	tbl.SetHeader(header.toRow())

	for i, pkg := range packages {
		// Discard packages where we cannot generate a sensible test summary.
		if pkg.NoTestFiles || pkg.NoTests || pkg.HasPanic {
			continue
		}
		pkgTests := getTestsFromPackages(pkg, option)
		all := make([]*parse.Test, 0, len(pkgTests.passed)+len(pkgTests.skipped))
		all = append(all, pkgTests.passed...)
		all = append(all, pkgTests.skipped...)

		for _, t := range all {
			// TODO(mf): why are we sorting this?
			t.SortEvents()

			testName := shortenTestName(t.Name, option.Trim)

			status := strings.ToUpper(t.Status().String())
			switch t.Status() {
			case parse.ActionPass:
				status = c.green(status)
			case parse.ActionSkip:
				status = c.yellow(status)
			case parse.ActionFail:
				status = c.red(status)
			}

			dir, packageName := path.Split(t.Package)
			// For SIV-style imports show the last non-versioned path identifer.
			// Example: github.com/foo/bar/helper/v3 returns helper/v3
			if dir != "" && versionMajorRe.MatchString(packageName) {
				_, subpath := path.Split(path.Clean(dir))
				packageName = path.Join(subpath, packageName)
			}
			row := testRow{
				status:      status,
				elapsed:     strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName:    testName,
				packageName: packageName,
			}
			tbl.Append(row.toRow())
		}
		if i != (len(packages) - 1) {
			// TODO(mf): is it possible to add a custom separator with tablewriter instead of empty space?
			tbl.Append(testRow{}.toRow())
		}
	}

	if tbl.NumLines() > 0 {
		// The table gets written to a strings builder so we can further modify the output
		// with lipgloss.
		tbl.Render()
		output := tableString.String()
		if c.format == OutputFormatBasic {
			output = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				Render(strings.TrimSuffix(output, "\n"))
		}
		fmt.Fprintln(c.w, output)
	}
}

func getTestsFromPackages(pkg *parse.Package, option TestTableOptions) *packageTests {
	tests := &packageTests{}
	if option.Skip {
		tests.skipped = append(tests.skipped, pkg.TestsByAction(parse.ActionSkip)...)
	}
	if option.Pass {
		tests.passed = append(tests.passed, pkg.TestsByAction(parse.ActionPass)...)

		// Order passed tests within a package by elapsed time DESC (longest on top).
		sort.Slice(tests.passed, func(i, j int) bool {
			return tests.passed[i].Elapsed() > tests.passed[j].Elapsed()
		})
		// Optionall, display only the slowest N tests by elapsed time.
		if option.Slow > 0 && len(tests.passed) > option.Slow {
			tests.passed = tests.passed[:option.Slow]
		}
	}
	return tests
}

func shortenTestName(s string, trim bool) string {
	var testName strings.Builder
	testName.WriteString(s)
	if trim && testName.Len() > 32 && strings.Count(testName.String(), "/") > 0 {
		testName.Reset()
		ss := strings.Split(s, "/")
		testName.WriteString(ss[0] + "\n")
		for i, s := range ss[1:] {
			testName.WriteString(" /" + s)
			if i != len(ss[1:])-1 {
				testName.WriteString("\n")
			}
		}
	}
	return testName.String()
}

func (c *consoleWriter) testsTableMarkdown(packages []*parse.Package, option TestTableOptions) {
	for _, pkg := range packages {
		// Print passed tests, sorted by elapsed DESC. Grouped by alphabetically sorted packages.
		var tableString strings.Builder
		tbl := newTableWriter(&tableString, c.format)

		header := []string{
			"Status",
			"Elapsed",
			"Test",
		}
		tbl.SetHeader(header)

		// Discard packages where we cannot generate a sensible test summary.
		if pkg.NoTestFiles || pkg.NoTests || pkg.HasPanic {
			continue
		}
		pkgTests := getTestsFromPackages(pkg, option)
		all := make([]*parse.Test, 0, len(pkgTests.passed)+len(pkgTests.skipped))
		all = append(all, pkgTests.passed...)
		all = append(all, pkgTests.skipped...)

		for _, t := range all {
			// TODO(mf): why are we sorting this?
			t.SortEvents()

			testName := shortenTestName(t.Name, option.Trim)

			status := strings.ToUpper(t.Status().String())
			switch t.Status() {
			case parse.ActionPass:
				status = c.green(status)
			case parse.ActionSkip:
				status = c.yellow(status)
			case parse.ActionFail:
				status = c.red(status)
			}
			tbl.Append([]string{
				status,
				strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName,
			})
		}
		if tbl.NumLines() > 0 {
			tbl.Render()

			fmt.Fprintf(c.w, "## 📦 Package **`%s`**\n", pkg.Summary.Package)
			fmt.Fprintln(c.w)
			fmt.Fprintf(c.w, "**%d passed** | **%d skipped**\n", len(pkgTests.passed), len(pkgTests.skipped))
			fmt.Fprintln(c.w)
			fmt.Fprintln(c.w, "<details>")
			fmt.Fprintln(c.w)
			fmt.Fprintln(c.w, "<summary>Click for test summary</summary>")
			fmt.Fprintln(c.w)
			fmt.Fprintln(c.w, tableString.String())
			fmt.Fprintln(c.w, "</details>")
			fmt.Fprintln(c.w)
		}
		fmt.Fprintln(c.w)
	}
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
