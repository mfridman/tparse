package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

var rawdump bytes.Buffer

// RawDump returns original lines of output from go test.
func RawDump() {
	sc := bufio.NewScanner(&rawdump)
	for sc.Scan() {
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			fmt.Fprintf(os.Stderr, "tparse new event error: %v\n", err)
			continue
		}
		fmt.Fprint(os.Stderr, e.Output)
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "tparse scan error: %v\n", err)
		return
	}
}

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

// Package is the representation of a single package being tested. The
// summary field is an event that contains all relevant information about the
// package, namely Package (name), Elapsed and Action (big pass or fail).
type Package struct {
	Summary *Event
	Tests   []*Test

	// NoTest indicates whether the package contains tests:
	// "?   \tpackage\t[no test files]\n"
	NoTest bool

	// Cached indicates whether the test result was obtained from the cache.
	Cached bool

	// Cover reports whether the package contains coverage (go test run with -cover)
	Cover    bool
	Coverage float64
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

	var scan bool
	var badLine int

	rd := io.TeeReader(r, &rawdump)

	sc := bufio.NewScanner(rd)
	for sc.Scan() {

		// We'll "prescan" up-to 50 lines for a parsable event. If we do get a parsable event
		// we expect no errors to follow until EOF. For each unparsable event we will dump back
		// the up to 50 lines to stderr.
		e, err := NewEvent(sc.Bytes())
		if err != nil {
			if scan || badLine >= 50 {
				return nil, err
			}

			fmt.Fprintln(os.Stderr, sc.Text())
			badLine++
			continue
		}
		scan = true

		pkg, ok := pkgs[e.Package]
		if !ok {
			pkg = &Package{Summary: &Event{}}
			pkgs[e.Package] = pkg
		}

		if e.SkipLine() {
			pkg.Summary = &Event{Action: ActionPass}
			pkg.NoTest = true
		}

		if e.IsCached() {
			pkg.Cached = true
		}

		cover, ok := e.Cover()
		if ok {
			pkg.Cover = true
			pkg.Coverage = cover
		}

		if e.Summary() {
			pkg.Summary = e
			continue
		}

		// We don't need to save these line.
		if e.Discard() {
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
