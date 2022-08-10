package app

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mfridman/tparse/parse"
	"github.com/olekukonko/tablewriter"
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
}

func (c *consoleWriter) summaryTable(packages []*parse.Package, showNoTests bool, options SummaryTableOptions) {
	var tableString strings.Builder
	tbl := newTableWriter(&tableString, c.format)
	tbl.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
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
	tbl.SetHeader(header.toRow())

	// Capture as separate slices because notests are optional when passed tests are available.
	// The only exception is if passed=0 and notests=1, then we display them regardless. This
	// is almost always the user matching on the wrong package.
	var passed, notests []summaryRow

	packagePrefix := findCommonPackagePrefix(packages)

	for _, pkg := range packages {
		elapsed := strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s"
		if pkg.Cached {
			elapsed = "(cached)"
		}
		packageName := shortenPackageName(pkg.Summary.Package, packagePrefix, 32, options.Trim)
		if pkg.HasPanic {
			row := summaryRow{
				status:      c.red("PANIC"),
				elapsed:     elapsed,
				packageName: packageName,
				cover:       "--", pass: "--", fail: "--", skip: "--",
			}
			tbl.Append(row.toRow())
			continue
		}
		if pkg.HasFailedBuildOrSetup {
			row := summaryRow{
				status:      c.red("FAIL"),
				elapsed:     elapsed,
				packageName: packageName + "\n[" + pkg.Summary.Output + "]",
				cover:       "--", pass: "--", fail: "--", skip: "--",
			}
			tbl.Append(row.toRow())
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

		coverage := "--"
		if pkg.Cover {
			coverage = fmt.Sprintf("%.1f%%", pkg.Coverage)
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

		status := strings.ToUpper(pkg.Summary.Action.String())
		switch pkg.Summary.Action {
		case parse.ActionPass:
			status = c.green(status)
		case parse.ActionSkip:
			status = c.yellow(status)
		case parse.ActionFail:
			status = c.red(status)
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

	for _, p := range passed {
		tbl.Append(p.toRow())
	}
	// Only display the "no tests to run" cases if users want to see them when passed
	// tests are available.
	// An exception is made if there are no passed tests and only a single no test files
	// package. This is almost always because the user forgot to match one or more packages.
	if showNoTests || (len(passed) == 0 && len(notests) == 1) {
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

func findCommonPackagePrefix(packages []*parse.Package) string {
	if len(packages) < 2 {
		return ""
	}

	prefixLength := 0
	for prefixLength = 0; prefixLength < len(packages[0].Summary.Package); prefixLength++ {
		for i := 0; i < len(packages); i++ {
			if len(packages[i].Summary.Package) == (prefixLength - 1) {
				goto End
			}
			if packages[0].Summary.Package[prefixLength] != packages[i].Summary.Package[prefixLength] {
				prefixLength--

				goto End
			}
		}
	}

End:
	if prefixLength <= 0 {
		return ""
	}

	prefix := packages[0].Summary.Package[0:prefixLength]
	lastSlash := strings.LastIndex(prefix, "/")
	if lastSlash >= 0 {
		prefix = prefix[0:lastSlash]
	}

	return prefix
}

func shortenPackageName(name string, prefix string, maxLength int, trim bool) string {
	if !trim {
		return name
	}

	if prefix == "" {
		dir, name := path.Split(name)
		// For SIV-style imports show the last non-versioned path identifer.
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
