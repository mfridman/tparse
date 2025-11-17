package app

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/mfridman/tparse/internal/utils"
	"github.com/mfridman/tparse/parse"
)

type SummaryTableOptions struct {
	// For narrow screens, remove common prefix and trim long package names vertically. Example:
	// github.com/mfridman/tparse/app
	// github.com/mfridman/tparse/internal/seed-up-down-to-zero
	//
	// tparse/app
	// tparse
	//  /seed-up-down-to-zero
	Trim bool

	// TrimPath is the path prefix to trim from the package name.
	TrimPath string

	// FailOnly will display only packages with failed tests.
	FailOnly bool
}

func (c *consoleWriter) summaryTable(
	packages []*parse.Package,
	showNoTests bool,
	options SummaryTableOptions,
	against *parse.GoTestSummary,
) {
	tbl := newTable(c.format, func(style lipgloss.Style, row, col int) lipgloss.Style {
		switch row {
		case table.HeaderRow:
		default:
			if col == 2 {
				// Package name
				style = style.Align(lipgloss.Left)
			}
		}
		return style
	})
	header := summaryRow{
		status:      "Status",
		elapsed:     "Elapsed",
		packageName: "Package",
		cover:       "Cover",
		pass:        "Pass",
		fail:        "Fail",
		skip:        "Skip",
	}
	tbl.Headers(header.toRow()...)
	data := table.NewStringData()

	// Capture as separate slices because notests are optional when passed tests are available.
	// The only exception is if passed=0 and notests=1, then we display them regardless. This
	// is almost always the user matching on the wrong package.
	var passed, notests []summaryRow

	names := make([]string, 0, len(packages))
	for _, pkg := range packages {
		names = append(names, pkg.Summary.Package)
	}
	packagePrefix := utils.FindLongestCommonPrefix(names)

	for _, pkg := range packages {
		elapsed := strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s"
		if pkg.Cached {
			elapsed = "(cached)"
		}
		packageName := pkg.Summary.Package
		packageName = shortenPackageName(packageName, packagePrefix, 32, options.Trim, options.TrimPath)
		if pkg.HasPanic {
			row := summaryRow{
				status:      c.red("PANIC"),
				elapsed:     elapsed,
				packageName: packageName,
				cover:       "--", pass: "--", fail: "--", skip: "--",
			}
			data.Append(row.toRow())
			continue
		}
		if pkg.HasFailedBuildOrSetup {
			row := summaryRow{
				status:      c.red("FAIL"),
				elapsed:     elapsed,
				packageName: packageName + "\n[" + pkg.Summary.Output + "]",
				cover:       "--", pass: "--", fail: "--", skip: "--",
			}
			data.Append(row.toRow())
			continue
		}
		if pkg.NoTestFiles {
			row := summaryRow{
				status:      c.yellow("NOTEST"),
				elapsed:     elapsed,
				packageName: packageName + "\n[no test files]",
				cover:       "--", pass: "--", fail: "--", skip: "--",
			}
			notests = append(notests, row)
			continue
		}
		if pkg.NoTests {
			// This should capture cases where packages truly have no tests, but empty files.
			if len(pkg.NoTestSlice) == 0 {
				row := summaryRow{
					status:      c.yellow("NOTEST"),
					elapsed:     elapsed,
					packageName: packageName + "\n[no tests to run]",
					cover:       "--", pass: "--", fail: "--", skip: "--",
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
				status:      c.yellow("NOTEST"),
				elapsed:     elapsed,
				packageName: packageName,
				cover:       "--", pass: "--", fail: "--", skip: "--",
			}
			notests = append(notests, row)

			if len(pkg.TestsByAction(parse.ActionPass)) == len(pkg.NoTestSlice) {
				continue
			}
		}
		// TODO(mf): refactor this
		// Separate cover colorization from the delta output.
		coverage := "--"
		if pkg.Cover {
			coverage = fmt.Sprintf("%.1f%%", pkg.Coverage)
			if against != nil {
				againstP, ok := against.Packages[pkg.Summary.Package]
				if ok {
					var sign string
					if pkg.Coverage > againstP.Coverage {
						sign = "+"
					}
					coverage = fmt.Sprintf("%s (%s)", coverage, sign+strconv.FormatFloat(pkg.Coverage-againstP.Coverage, 'f', 1, 64)+"%")
				} else {
					coverage = fmt.Sprintf("%s (-)", coverage)
				}
			}
			// Showing coverage for a package that failed is a bit odd.
			//
			// Only colorize the coverage when everything passed AND the output is not markdown.
			if pkg.Summary.Action == parse.ActionPass && c.format != OutputFormatMarkdown {
				switch cover := pkg.Coverage; {
				case cover > 0.0 && cover <= 50.0:
					coverage = c.red(coverage)
				case pkg.Coverage > 50.0 && pkg.Coverage < 80.0:
					coverage = c.yellow(coverage)
				case pkg.Coverage >= 80.0:
					coverage = c.green(coverage)
				}
			}
		}

		status := c.FormatAction(pkg.Summary.Action)

		// Skip packages with no coverage to mimic nocoverageredesign behavior (changed in github.com/golang/go/issues/24570)
		totalTests := len(pkg.TestsByAction(parse.ActionPass)) + len(pkg.TestsByAction(parse.ActionFail)) + len(pkg.TestsByAction(parse.ActionSkip))
		if pkg.Cover && pkg.Coverage == 0.0 && totalTests == 0 {
			continue
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

	if data.Rows() == 0 && len(passed) == 0 && len(notests) == 0 {
		return
	}
	for _, r := range passed {
		if options.FailOnly && r.fail == "0" {
			continue
		}
		data.Append(r.toRow())
	}

	// Only display the "no tests to run" cases if users want to see them when passed
	// tests are available.
	// An exception is made if there are no passed tests and only a single no test files
	// package. This is almost always because the user forgot to match one or more packages.
	if showNoTests || (len(passed) == 0 && len(notests) == 1) {
		for _, r := range notests {
			data.Append(r.toRow())
		}
	}
	if options.FailOnly && data.Rows() == 0 {
		fmt.Fprintln(c, "No tests failed.")
		return
	}
	fmt.Fprintln(c, tbl.Data(data).Render())
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

func shortenPackageName(
	name string,
	prefix string,
	maxLength int,
	trim bool,
	trimPath string,
) string {
	if trimPath == "auto" {
		name = strings.TrimPrefix(name, prefix)
	} else if trimPath != "" {
		name = strings.TrimPrefix(name, trimPath)
	}
	if !trim {
		return name
	}

	if prefix == "" {
		dir, name := path.Split(name)
		// For SIV-style imports show the last non-versioned path identifier.
		// Example: github.com/foo/bar/helper/v3 returns helper/v3
		if dir != "" && versionMajorRe.MatchString(name) {
			_, subpath := path.Split(path.Clean(dir))
			name = path.Join(subpath, name)
		}
		return name
	}

	name = strings.TrimPrefix(name, prefix)
	name = strings.TrimLeft(name, "/")
	name = shortenTestName(name, true, maxLength)

	return name
}
