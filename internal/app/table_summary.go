package app

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mfridman/tparse/parse"
)

func (c *consoleWriter) summaryTable(packages parse.Packages, showNoTests bool) {
	var tableString strings.Builder
	tbl := newTableWriter(&tableString, c.format)

	header := summaryRow{
		status:      "Status",
		elapsed:     "Elapsed",
		packageName: "Package",
		cover:       "Cover",
		pass:        "Pass",
		fail:        "Fail",
		skip:        "Skip",
	}
	tbl.SetHeader(header.toRow())

	var passed, notests []summaryRow

	for packageName, pkg := range packages {
		elapsed := strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s"
		if pkg.Cached {
			elapsed = "(cached)"
		}
		if pkg.HasPanic {
			row := summaryRow{
				status:      c.red("PANIC", true),
				elapsed:     elapsed,
				packageName: packageName,
				cover:       "--",
				pass:        "--",
				fail:        "--",
				skip:        "--",
			}
			tbl.Append(row.toRow())
			continue
		}
		if pkg.NoTestFiles {
			row := summaryRow{
				status:      c.yellow("NOTEST", true),
				elapsed:     elapsed,
				packageName: packageName + "\n[no test files]",
				cover:       "--",
				pass:        "--",
				fail:        "--",
				skip:        "--",
			}
			notests = append(notests, row)
			continue
		}
		if pkg.NoTests {
			// This should capture cases where packages truly have no tests, but empty files.
			if len(pkg.NoTestSlice) == 0 {
				row := summaryRow{
					status:      c.yellow("NOTEST", true),
					elapsed:     elapsed,
					packageName: packageName + "\n[no tests to run]",
					cover:       "--",
					pass:        "--",
					fail:        "--",
					skip:        "--",
				}
				notests = append(notests, row)
				continue
			}
			// This should capture cases where packages have a mixture of empty and non-empty test files.
			var ss []string
			for i, t := range pkg.NoTestSlice {
				i++
				ss = append(ss, fmt.Sprintf("%d.%s", i, t.Test))
			}
			packageName := fmt.Sprintf("%s\n[no tests to run]\n%s", packageName, strings.Join(ss, "\n"))
			row := summaryRow{
				status:      c.yellow("NOTEST", true),
				elapsed:     elapsed,
				packageName: packageName,
				cover:       "--",
				pass:        "--",
				fail:        "--",
				skip:        "--",
			}
			notests = append(notests, row)

			if len(pkg.TestsByAction(parse.ActionPass)) == len(pkg.NoTestSlice) {
				continue
			}
		}

		coverage := fmt.Sprintf("%.1f%%", pkg.Coverage)
		if pkg.Summary.Action != parse.ActionFail {
			switch cover := pkg.Coverage; {
			case cover > 0.0 && cover <= 50.0:
				coverage = c.red(coverage, false)
			case pkg.Coverage > 50.0 && pkg.Coverage < 80.0:
				coverage = c.yellow(coverage, false)
			case pkg.Coverage >= 80.0:
				coverage = c.green(coverage, false)
			}
		}

		status := strings.ToUpper(pkg.Summary.Action.String())
		switch pkg.Summary.Action {
		case parse.ActionPass:
			status = c.green(status, false)
		case parse.ActionSkip:
			status = c.yellow(status, false)
		case parse.ActionFail:
			status = c.red(status, false)
		}
		row := summaryRow{
			status:      status,
			elapsed:     elapsed,
			packageName: packageName,
			cover:       coverage,
			pass:        strconv.Itoa(len(pkg.TestsByAction(parse.ActionPass))),
			fail:        strconv.Itoa(len(pkg.TestsByAction(parse.ActionFail))),
			skip:        strconv.Itoa(len(pkg.TestsByAction(parse.ActionSkip))),
		}
		passed = append(passed, row)
	}

	if tbl.NumLines() == 0 && len(passed) == 0 && len(notests) == 0 {
		return
	}
	// Sort package tests by name ASC.
	// TODO(mf): what about sorting by elapsed, probably DESC, to quickly gauge
	// slow running tests? Too many knobs makes this tool more complicated to use.
	sortSummaryRows(passed, ASC)
	sortSummaryRows(notests, ASC)

	for _, p := range passed {
		tbl.Append(p.toRow())
	}
	// Only display the "no tests to run" cases if users want to see them when passed
	// tests are available.
	if showNoTests {
		for _, p := range notests {
			tbl.Append(p.toRow())
		}
	}
	// The table gets written to a strings builder so we can further modify the output
	// with lipgloss.
	tbl.Render()
	output := tableString.String()
	if c.format == OutputFormatBasic {
		output = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Render(strings.TrimSuffix(tableString.String(), "\n"))
	}
	fmt.Fprintln(c.w, output)
}

type summaryRow struct {
	status      string
	elapsed     string
	packageName string
	cover       string
	pass        string
	fail        string
	skip        string
}

func (r summaryRow) toRow() []string {
	return []string{
		r.status,
		r.elapsed,
		r.packageName,
		r.cover,
		r.pass,
		r.fail,
		r.skip,
	}
}

type orderBy int

const (
	ASC orderBy = iota + 1
	DESC
)

func sortSummaryRows(rows []summaryRow, order orderBy) {
	sort.Slice(rows, func(i, j int) bool {
		if order == ASC {
			return rows[i].packageName < rows[j].packageName
		}
		return rows[i].packageName > rows[j].packageName
	})
}
