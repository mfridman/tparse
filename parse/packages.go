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
		"Status",
		"Elapsed",
		"Package",
		"Cover",
		"Pass",
		"Fail",
		"Skip",
	})

	for name, pkg := range p {

		if pkg.NoTest {
			if !skipNoTests {
				continue
			}

			tbl.Append([]string{
				Yellow("SKIP"),
				"0.00s",
				name + "\n[no test files]",
				fmt.Sprintf(" %.1f%%", pkg.Coverage),
				"0", "0", "0",
			})

			continue
		}

		if pkg.Cached {
			name += " (cached)"
		}

		tbl.Append([]string{
			pkg.Summary.Action.WithColor(),
			strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s",
			name,
			fmt.Sprintf(" %.1f%%", pkg.Coverage),
			strconv.Itoa(len(pkg.TestsByAction(ActionPass))),
			strconv.Itoa(len(pkg.TestsByAction(ActionFail))),
			strconv.Itoa(len(pkg.TestsByAction(ActionSkip))),
		})
	}

	tbl.Render()
	fmt.Printf("\n")
}
