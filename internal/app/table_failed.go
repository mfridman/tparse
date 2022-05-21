package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mfridman/tparse/parse"
)

// printFailed prints all failed tests, grouping them by package. Packages are sorted.
// Panic is an exception.
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
			// TODO(mf): document why panics are handled separately. A panic may or may
			// not be associated with tests, so we print it at the package level.
			output := prepareStyledPanic(pkg.Summary.Package, pkg.Summary.Test, pkg.PanicEvents)
			fmt.Fprintln(c.w, output)
			continue
		}
		failedTests := pkg.TestsByAction(parse.ActionFail)
		if len(failedTests) == 0 {
			continue
		}

		styledPackageHeader := styledHeader(
			strings.ToUpper(pkg.Summary.Action.String()),
			pkg.Summary.Package,
		)
		// TODO(mf): should the tests be sorted, probably alphabetically ASC?
		styledTestStrings := make([]string, 0, len(failedTests))
		for i, t := range failedTests {
			// Add bottom border to all tests except the last one.
			addBorder := (i != len(failedTests)-1)
			styledTestStrings = append(styledTestStrings, prepareStyledTest(t, addBorder))
		}
		combined := append(
			[]string{styledPackageHeader}, // Package header (1 row)
			styledTestStrings...,          // Tests (multiple rows)
		)
		fmt.Fprintln(c.w, lipgloss.JoinVertical(lipgloss.Left, combined...))
	}
}

func prepareStyledPanic(packageName, testName string, panicEvents []*parse.Event) string {
	styledPackageHeader := styledHeader(
		"PANIC",
		packageName,
	)
	// TODO(mf): can we pass this panic stack to another package and either by default,
	// or optionally, build human-readable panic output with:
	// https://github.com/maruel/panicparse
	var rows strings.Builder
	for _, e := range panicEvents {
		if e.Output == "" {
			continue
		}
		rows.WriteString(e.Output)
	}
	return lipgloss.JoinVertical(lipgloss.Left, styledPackageHeader, rows.String())
}

// styledHeader styles a header based on the status and package name:
//
// ╭───────────────────────────────────────────────────────────╮
// │   PANIC  package: github.com/pressly/goose/v3/tests/e2e   │
// ╰───────────────────────────────────────────────────────────╯
//
func styledHeader(status, packageName string) string {
	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("103"))
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true).
		PaddingLeft(3).
		PaddingRight(2)
	packageNameStyle := lipgloss.NewStyle().
		PaddingRight(3)
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusStyle.Render(status),
		packageNameStyle.Render("package: "+packageName),
	)
	return headerStyle.Render(headerRow)
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
