package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

// Packages is a collection of packages being tested.
type Packages map[string]*Package

func (p Packages) PrintSummary(skipNoTests bool) {
	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.SetHeader([]string{
		"Status",  //0
		"Elapsed", //1
		"Package", //2
		"Cover",   //3
		"Pass",    //4
		"Fail",    //5
		"Skip",    //6
	})

	for name, pkg := range p {

		if pkg.NoTest {
			if skipNoTests {
				tbl.Append([]string{
					colorize("SKIP", cYellow, true),
					"--",
					name + "\n[no test files]",
					fmt.Sprintf(" %.1f%%", pkg.Coverage),
					"--", "--", "--",
				})
			}
			continue
		}

		var elapsed string
		if pkg.Cached {
			elapsed = "(cached)"
		} else {
			elapsed = strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s"
		}

		coverage := fmt.Sprintf("%.1f%%", pkg.Coverage)
		switch c := pkg.Coverage; {
		case c == 0.0:
			break
		case c <= 50.0:
			coverage = colorize(coverage, cRed, true)
		case pkg.Coverage > 50.0 && pkg.Coverage < 80.0:
			coverage = colorize(coverage, cYellow, true)
		case pkg.Coverage >= 80.0:
			coverage = colorize(coverage, cGreen, true)
		}

		tbl.Append([]string{
			pkg.Summary.Action.WithColor(), //0
			elapsed,                        //1
			name,                           //2
			coverage,                       //3
			strconv.Itoa(len(pkg.TestsByAction(ActionPass))), //4
			strconv.Itoa(len(pkg.TestsByAction(ActionFail))), //5
			strconv.Itoa(len(pkg.TestsByAction(ActionSkip))), //6
		})
	}

	if tbl.NumLines() > 0 {
		tbl.Render()
		fmt.Printf("\n")
	} else {
		RawDump()
	}
}

func (p Packages) PrintFailed() {
	// Print all failed tests per package (if any).
	for _, pkg := range p {
		failed := pkg.TestsByAction(ActionFail)
		if len(failed) == 0 {
			continue
		}

		s := fmt.Sprintf("PACKAGE: %s", pkg.Summary.Package)
		n := make([]string, len(s)+1)
		fmt.Printf("%s\n%s\n", s, strings.Join(n, "-"))

		for _, t := range failed {
			t.Sort()

			fmt.Printf("%s\n\n", t.Stack())

		}
	}
}

func (p Packages) PrintTests(pass, skip, trim bool) {
	// Print passed tests, sorted by elapsed. Unlike failed tests, passed tests
	// are not grouped. Maybe bad design?
	tbl := tablewriter.NewWriter(os.Stdout)

	tbl.SetHeader([]string{
		"Status",
		"Elapsed",
		"Test",
		"Package",
	})

	tbl.SetAutoWrapText(false)

	var i int
	for _, pkg := range p {
		var all []*Test
		if skip {
			skipped := pkg.TestsByAction(ActionSkip)
			all = append(all, skipped...)
		}
		if pass {
			passed := pkg.TestsByAction(ActionPass)

			// Sort tests within a package by elapsed time in descending order, longest on top.
			sort.Slice(passed, func(i, j int) bool {
				return passed[i].Elapsed() > passed[j].Elapsed()
			})

			all = append(all, passed...)
		}
		if len(all) == 0 {
			continue
		}

		for _, t := range all {
			t.Sort()

			var testName strings.Builder
			testName.WriteString(t.Name)
			if trim && testName.Len() > 32 && strings.Count(testName.String(), "/") > 0 {
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

			tbl.Append([]string{
				t.Status().WithColor(),
				strconv.FormatFloat(t.Elapsed(), 'f', 2, 64),
				testName.String(),
				filepath.Base(t.Package),
			})
		}

		// Add empty line between package groups except last one.
		if i != len(p)-1 {
			tbl.Append([]string{"", "", "", ""})
		}
		i++
	}
	if tbl.NumLines() > 0 {
		tbl.Render()
	}
}
