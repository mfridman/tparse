package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mfridman/tparse/parse"
)

// printFailed prints all failed tests grouing them by package . Panic is an exception.
func (c *consoleWriter) printFailed(packages parse.Packages) {
	sortedPackages := make([]*parse.Package, 0, len(packages))
	for _, pkg := range packages {
		sortedPackages = append(sortedPackages, pkg)
	}
	// Sort packages alphabetically.
	sort.Slice(sortedPackages, func(i, j int) bool {
		return sortedPackages[i].Summary.Package < sortedPackages[j].Summary.Package
	})
	for _, pkg := range sortedPackages {
		if pkg.HasPanic {
			// may or may not be associated with tests, so we print it separately.
			// w.PrintPanic(pkg)
			// TODO(mf): implement special panic print logic.
			continue
		}
		failedTests := pkg.TestsByAction(parse.ActionFail)
		if len(failedTests) == 0 {
			continue
		}

		headerStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("60"))
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			PaddingLeft(3).
			PaddingRight(2)
		packageNameStyle := lipgloss.NewStyle().
			PaddingRight(3)
		headerRow := lipgloss.JoinHorizontal(
			lipgloss.Left,
			statusStyle.Render(strings.ToUpper(pkg.Summary.Action.String())),
			packageNameStyle.Render("package: "+pkg.Summary.Package),
		)

		// TODO(mf): should the tests be sorted, probably alphabetically ASC?
		styledTestStrings := make([]string, 0, len(failedTests))
		for i, t := range failedTests {
			// Add bottom border to all tests except the last one.
			addBorder := (i != len(failedTests)-1)
			styledTestStrings = append(styledTestStrings, prepareStyledTest(t, addBorder))
		}
		combined := append([]string{headerStyle.Render(headerRow)}, styledTestStrings...)
		fmt.Fprintln(c.w, lipgloss.JoinVertical(lipgloss.Left, combined...))
	}
}

func prepareStyledTest(t *parse.Test, bottomBorder bool) string {
	t.SortEvents()

	var rows, headerRows strings.Builder
	for _, e := range t.Events {
		// Only add events that have output information. Skip everything else.
		// Note, since we know about all the output, we can bubble "--- Fail" to the top
		// of the output so it's trivial to spot the failing test name and elapsed time.
		if e.Action != parse.ActionOutput {
			continue
		}
		if strings.HasPrefix(e.Output, "--- FAIL: ") {
			header := lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Render(e.Output)
			headerRows.WriteString(header)
			continue
		}
		rows.WriteString(e.Output)
	}
	combined := lipgloss.JoinVertical(
		lipgloss.Left,
		headerRows.String(),
		rows.String(),
	)
	border := lipgloss.NormalBorder()
	if !bottomBorder {
		border = lipgloss.HiddenBorder()
	}
	return lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(border).
		Render(combined)
}
