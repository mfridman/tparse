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
			output := c.prepareStyledPanic(pkg.Summary.Package, pkg.Summary.Test, pkg.PanicEvents)
			fmt.Fprintln(c.w, output)
			continue
		}
		failedTests := pkg.TestsByAction(parse.ActionFail)
		if len(failedTests) == 0 {
			continue
		}
		styledPackageHeader := c.styledHeader(
			pkg.Summary.Action.String(),
			pkg.Summary.Package,
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
			Faint(c.format != OutputFormatMarkdown).
			Width(96)

		/*
			Note, some output such as the "--- FAIL: " line is prefixed
			with spaces. Unfortunately when dumping this in markdown format
			it renders as an code block.

			"To produce a code block in Markdown, simply indent every line of the
			block by at least 4 spaces or 1 tab."
			Ref. https://daringfireball.net/projects/markdown/syntax

			Example:
			 --- FAIL: Test (0.05s)
			    --- FAIL: Test/test_01 (0.01s)
			        --- FAIL: Test/test_01/sort (0.00s)

			This is why we wrap the entire test output in a code block.
		*/

		if c.format == OutputFormatMarkdown {
			fmt.Fprintln(c.w, fencedCodeBlock)
		}
		var key string
		for i, t := range failedTests {
			// Add top divider to all tests except first one.
			base, _, _ := cut(t.Name, "/")
			if i > 0 && key != base {
				fmt.Fprintln(c.w, divider.String())
			}
			key = base
			fmt.Fprintln(c.w, c.prepareStyledTest(t))
		}
		if c.format == OutputFormatMarkdown {
			fmt.Fprint(c.w, fencedCodeBlock+"\n\n")
		}
	}
}

const (
	fencedCodeBlock string = "```"
)

// copied directly from strings.Cut (go1.18) to support older Go versions.
// In the future, replace this with the upstream function.
func cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

func (c *consoleWriter) prepareStyledPanic(
	packageName string,
	testName string,
	panicEvents []*parse.Event,
) string {
	if testName != "" {
		packageName = packageName + " • " + testName
	}
	styledPackageHeader := c.styledHeader("PANIC", packageName)
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

func (c *consoleWriter) styledHeader(status, packageName string) string {
	status = c.red(strings.ToUpper(status))
	packageName = strings.TrimSpace(packageName)

	if c.format == OutputFormatMarkdown {
		msg := fmt.Sprintf("## %s • %s", status, packageName)
		return msg
		// TODO(mf): an alternative implementation is to add 2 horizontal lines above and below
		// the package header output.
		//
		// var divider string
		// for i := 0; i < len(msg); i++ {
		// 	divider += "─"
		// }
		// return fmt.Sprintf("%s\n%s\n%s", divider, msg, divider)
	}
	/*
		Need to rethink how to best support multiple output formats across
		CI, local terminal development and markdown

		See https://github.com/mfridman/tparse/issues/71
	*/
	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("103"))
	statusStyle := lipgloss.NewStyle().
		PaddingLeft(3).
		PaddingRight(2).
		Foreground(lipgloss.Color("9"))
	packageNameStyle := lipgloss.NewStyle().
		PaddingRight(3)
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusStyle.Render(status),
		packageNameStyle.Render("package: "+packageName),
	)
	return headerStyle.Render(headerRow)
}

const (
	failLine = "--- FAIL: "
)

func (c *consoleWriter) prepareStyledTest(t *parse.Test) string {
	t.SortEvents()

	var rows, headerRows strings.Builder
	for _, e := range t.Events {
		// Only add events that have output information. Skip everything else.
		// Note, since we know about all the output, we can bubble "--- Fail" to the top
		// of the output so it's trivial to spot the failing test name and elapsed time.
		if e.Action != parse.ActionOutput {
			continue
		}
		if strings.Contains(e.Output, failLine) {
			header := strings.TrimSuffix(e.Output, "\n")
			// go test prefixes too much padding to the "--- FAIL: " output lines.
			// Let's cut the padding by half, being careful to preserve the fail
			// line and the proceeding output.
			before, after, ok := cut(header, failLine)
			var pad string
			if ok {
				var n int
				for _, r := range before {
					if r == 32 {
						n++
					}
				}
				for i := 0; i < n/2; i++ {
					pad += " "
				}
			}
			header = pad + failLine + after

			// Avoid colorizing markdown output so it renders properly, otherwise add a subtle
			// red color to the test headers.
			if c.format != OutputFormatMarkdown {
				header = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(header)
			}
			headerRows.WriteString(header)
			continue
		}

		if e.Output != "" {
			rows.WriteString(e.Output)
		}
	}
	out := headerRows.String()
	if rows.Len() > 0 {
		out += "\n\n" + rows.String()
	}
	return out
}
