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
func (c *consoleWriter) printFailed(packages []*parse.Package) {
	for _, pkg := range packages {
		if pkg.HasPanic {
			// TODO(mf): document why panics are handled separately. A panic may or may
			// not be associated with tests, so we print it at the package level.
			output := prepareStyledPanic(pkg.Summary.Package, pkg.Summary.Test, pkg.PanicEvents, c.disableColor)
			fmt.Fprintln(c.w, output)
			continue
		}
		failedTests := pkg.TestsByAction(parse.ActionFail)
		if len(failedTests) == 0 {
			continue
		}

		styledPackageHeader := styledHeader(
			strings.ToUpper(pkg.Summary.Action.String()),
			strings.TrimSpace(pkg.Summary.Package),
			c.disableColor,
		)
		fmt.Fprintln(c.w, styledPackageHeader)
		fmt.Fprintln(c.w)
		/*
			Failed tests are all the individual tests, where the subtests are not separated.

			We need to sort the tests by name to ensure they are grouped together
		*/
		sort.Slice(failedTests, func(i, j int) bool {
			return failedTests[i].Name < failedTests[j].Name
		})

		divider := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			Faint(true).
			Width(96)

		var key string
		for i, t := range failedTests {
			// Add top divider to all tests except first one.
			base, _, _ := cut(t.Name, "/")
			if i > 0 && key != base {
				fmt.Fprintln(c.w, divider.String())
			}
			key = base
			fmt.Fprintln(c.w, prepareStyledTest(t))
		}
	}
}

// copied directly from strings.Cut (go1.18) to support older Go versions.
// In the future, replace this with the upstream function.
func cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

func prepareStyledPanic(packageName, testName string, panicEvents []*parse.Event, disableColor bool) string {
	if testName != "" {
		packageName = packageName + " • " + testName
	}
	styledPackageHeader := styledHeader(
		"PANIC",
		packageName,
		disableColor,
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
func styledHeader(status, packageName string, disableColor bool) string {
	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("103"))
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
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

func prepareStyledTest(t *parse.Test) string {
	t.SortEvents()

	var rows, headerRows strings.Builder
	for _, e := range t.Events {
		// Only add events that have output information. Skip everything else.
		// Note, since we know about all the output, we can bubble "--- Fail" to the top
		// of the output so it's trivial to spot the failing test name and elapsed time.
		if e.Action != parse.ActionOutput {
			continue
		}
		if strings.Contains(e.Output, "--- FAIL: ") {
			header := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(e.Output)
			headerRows.WriteString(header)
			continue
		}
		if e.Output != "" {
			rows.WriteString(e.Output)
		}
	}
	out := headerRows.String()
	if rows.Len() > 0 {
		out += "\n" + rows.String()
	}
	return out
}
