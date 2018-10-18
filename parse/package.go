package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// Packages is a collection of packages being tested.
// TODO: consider changing this to a slice of packages instead of a map?
// - would make it easier sorting the summary box by elapsed time
type Packages map[string]*Package

func (p Packages) Print() {
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
		tbl.Append([]string{
			pkg.Summary.Action.WithColor(),
			strconv.FormatFloat(pkg.Summary.Elapsed, 'f', 2, 64),
			name,
			strconv.Itoa(len(pkg.Passed())),
			strconv.Itoa(len(pkg.Failed())),
			strconv.Itoa(len(pkg.Skipped())),
		})
	}

	tbl.Render()
	fmt.Printf("\n")
}

// Package is the representation of a single package being tested.
type Package struct {
	Summary *Event // single summary event describing the result of the package tests(s)
	Tests   []*Test
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

func Do(r io.Reader) (Packages, error) {

	pkgs := Packages{}

	sc := bufio.NewScanner(r)
	for sc.Scan() {

		e, err := NewEvent(bytes.NewReader(sc.Bytes()))
		if err != nil {
			// TODO(mf): consider logging? and continue scanning instead of failing.
			return nil, err
		}

		if e.Discard() {
			continue
		}

		pkg, ok := pkgs[e.Package]
		if !ok {
			pkg = &Package{}
			pkgs[e.Package] = pkg
		}

		if e.IsSummary() {
			pkg.Summary = e
			continue
		}

		pkg.AddTestEvent(e)

	}

	if err := sc.Err(); err != nil {
		// something went wrong scanning. We may want to fail? and dump
		// what we were able to read.
		// E.g., store events in strings.Builder and dump the output lines,
		// or return a structured error with context and events we were able to read.
		return nil, errors.Wrap(err, "bufio scanner error")
	}

	return pkgs, nil
}

// Passed returns a slice of tests, sorted by time.
func (p *Package) Passed() []*Test {
	passed := []*Test{}

	for _, t := range p.Tests {
		if t.Status() == ActionPass {
			passed = append(passed, t)
		}
	}
	if len(passed) == 0 {
		return passed
	}

	sort.Slice(passed, func(i, j int) bool {
		return passed[i].Elapsed() > passed[i].Elapsed()
	})

	return passed
}

func (p *Package) Failed() []*Test {
	failed := []*Test{}

	for _, t := range p.Tests {
		if t.Status() == ActionFail {
			failed = append(failed, t)
		}
	}

	return failed
}

func (p *Package) Skipped() []*Test {
	skipped := []*Test{}

	for _, t := range p.Tests {
		if t.Status() == ActionSkip {
			skipped = append(skipped, t)
		}
	}

	return skipped
}
