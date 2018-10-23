package parse

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// Packages is a collection of packages being tested.
// TODO: consider changing this to a slice of packages instead of a map?
// - would make it easier sorting the summary box by elapsed time
// - would make it easier adding functional options.
type Packages map[string]*Package

func (p Packages) Print(skipNoTests bool) {
	if len(p) == 0 {
		return
	}

	tbl := tablewriter.NewWriter(os.Stdout)

	tbl.SetHeader([]string{
		"Status",
		"Elapsed",
		"Package",
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
				"0", "0", "0",
			})

			continue
		}

		tbl.Append([]string{
			pkg.Summary.Action.WithColor(),
			strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64) + "s",
			name,
			strconv.Itoa(len(pkg.TestsByAction(ActionPass))),
			strconv.Itoa(len(pkg.TestsByAction(ActionFail))),
			strconv.Itoa(len(pkg.TestsByAction(ActionSkip))),
		})
	}

	tbl.Render()
	fmt.Printf("\n")
}

// Package is the representation of a single package being tested. The
// summary field is an event that contains all relevant information about the
// package, namely Package (name), Elapsed and Action (big pass or fail).
type Package struct {
	Summary *Event
	Tests   []*Test

	// NoTest indicates whether the package contains tests:
	// "?   \tpackage\t[no test files]\n"
	NoTest bool
}

// AddTestEvent adds the event to a test based on test name.
func (p *Package) AddTestEvent(event *Event) {
	for _, t := range p.Tests {
		if t.Name == event.Test {
			t.Events = append(t.Events, event)
			return
		}
	}

	t := &Test{
		Name:    event.Test,
		Package: event.Package,
	}
	t.Events = append(t.Events, event)

	p.Tests = append(p.Tests, t)
}

// Start will return PanicErr on the first package that reports a test containing a panic.
func Start(r io.Reader) (Packages, error) {

	pkgs := Packages{}

	sc := bufio.NewScanner(r)
	for sc.Scan() {

		e, err := NewEvent(sc.Bytes())
		if err != nil {
			return nil, err
		}

		if e.Discard() {
			continue
		}

		pkg, ok := pkgs[e.Package]
		if !ok {
			pkg = &Package{Summary: &Event{}}
			pkgs[e.Package] = pkg
		}

		if e.SkipLine() {
			pkg.Summary = &Event{Action: ActionPass}
			pkg.NoTest = true
		}

		if e.Summary() {
			pkg.Summary = e
			continue
		}

		pkg.AddTestEvent(e)

	}

	if err := sc.Err(); err != nil {
		// TODO: FIXME: something went wrong scanning. We may want to fail? and dump
		// what we were able to read.
		// E.g., store events in strings.Builder and dump the output lines,
		// or return a structured error with context and events we were able to read.
		return nil, errors.Wrap(err, "bufio scanner error")
	}

	// Panic means end of the world, don't return a summary, no table tests, etc.
	// Just return the package, test and debug info from the panic inside the error.
	for _, pkg := range pkgs {
		if err := pkg.HasPanic(); err != nil {
			return nil, err
		}
	}

	return pkgs, nil
}

// HasPanic reports whether a package contains a test that panicked.
// A PanicErr is returned if a test contains a panic.
func (p *Package) HasPanic() error {
	for _, t := range p.Tests {
		for i := range t.Events {
			if strings.HasPrefix(t.Events[i].Output, "panic:") && strings.HasPrefix(t.Events[i+1].Output, "\tpanic:") {
				return &PanicErr{Test: t, Summary: t.Events[len(t.Events)-1]}
			}
		}
	}

	return nil
}

// TestsByAction returns all tests that identify as one of the following
// actions: pass, skip or fail.
//
// An empty slice if returned if there are no tests.
func (p *Package) TestsByAction(action Action) []*Test {
	tests := []*Test{}

	for _, t := range p.Tests {
		if t.Status() == action {
			tests = append(tests, t)
		}
	}

	return tests
}
