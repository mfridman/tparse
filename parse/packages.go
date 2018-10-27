package parse

import (
	"fmt"
	"os"
	"strconv"

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

	tbl.Render()
	fmt.Printf("\n")
}
