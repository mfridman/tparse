package app

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mfridman/tparse/parse"
	"github.com/olekukonko/tablewriter"
)

type testTableOptions struct {
	pass, skip, trim bool
	slow             int
}

func (c *consoleWriter) TestsTable(packages parse.Packages, option testTableOptions) {
	// Print passed tests, sorted by elapsed DESC. Grouped by alphabetically sorted packages.
	tbl := tablewriter.NewWriter(c.w)

	header := testRow{
		status:      "Status",
		elapsed:     "Elapsed",
		testName:    "Test",
		packageName: "Package",
	}
	tbl.SetHeader(header.toRow())

	switch c.format {
	case OutputFormatPlain:
		tbl.SetBorder(false)
		tbl.SetRowSeparator("")
		tbl.SetColumnSeparator("")
		tbl.SetHeaderLine(false)
	case OutputFormatMarkdown:
		tbl.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		tbl.SetCenterSeparator("|")
	}
	tbl.SetAutoWrapText(false)

	// Sort packages alphabetically by name ASC.
	var packageNames []string
	for name, _ := range packages {
		packageNames = append(packageNames, name)
	}
	sort.Strings(packageNames)

	for i, name := range packageNames {
		pkg := packages[name]
		// Discard packages where we cannot generate a sensible test summary.
		if pkg.NoTestFiles || pkg.NoTests || pkg.HasPanic {
			continue
		}
		var skipped, passed []*parse.Test

		if option.skip {
			skipped = append(skipped, pkg.TestsByAction(parse.ActionSkip)...)
		}
		if option.pass {
			passed = append(passed, pkg.TestsByAction(parse.ActionPass)...)

			// Order passed tests within a package by elapsed time DESC (longest on top).
			sort.Slice(passed, func(i, j int) bool {
				return passed[i].Elapsed() > passed[j].Elapsed()
			})
			// Optionall, display only the slowest N tests by elapsed time.
			if option.slow > 0 && len(passed) > option.slow {
				passed = passed[:option.slow]
			}
		}

		all := make([]*parse.Test, 0, len(skipped)+len(passed))
		all = append(all, skipped...)
		all = append(all, passed...)

		for _, t := range all {
			// TODO(mf): why are we sorting this?
			t.SortEvents()

			var testName strings.Builder
			testName.WriteString(t.Name)
			if option.trim && testName.Len() > 32 && strings.Count(testName.String(), "/") > 0 {
				testName.Reset()
				ss := strings.Split(t.Name, "/")
				testName.WriteString(ss[0] + "\n")
				for i, s := range ss[1:] {
					testName.WriteString(" /" + s)
					if i != len(ss[1:])-1 {
						testName.WriteString("\n")
					}
				}
			}
			status := t.Status().String()
			switch pkg.Summary.Action {
			case parse.ActionPass:
				status = c.green(status, false)
			case parse.ActionSkip:
				status = c.yellow(status, false)
			case parse.ActionFail:
				status = c.red(status, false)
			}
			r := regexp.MustCompile(`(?m)v[0-9]+`)
			dir, packageName := path.Split(t.Package)
			// For SIV-style imports show the last non-versioned path identifer.
			// Example: "github.com/foo/bar/helper/v1" return helper/v3
			if dir != "" && r.MatchString(packageName) {
				_, subpath := path.Split(path.Clean(dir))
				packageName = path.Join(subpath, packageName)
			}
			row := testRow{
				status:      status,
				elapsed:     strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName:    testName.String(),
				packageName: packageName,
			}

			tbl.Append(row.toRow())
		}
		// Add empty line between package groups except the last package
		if i+1 < len(sp) {
			tbl.Append(testRow{}.toRow())
		}
	}

	if tbl.NumLines() > 0 {
		fmt.Fprintln(c.w)
		tbl.Render()
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
