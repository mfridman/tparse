package parse

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type PanicErr struct {
	Summary *Event
	Test    *Test
}

var _ error = (*PanicErr)(nil)

func (p *PanicErr) Error() string {
	return fmt.Sprintf("panic: package: %s: test: %s", p.Summary.Package, p.Test.Name)
}

func (p *PanicErr) PrintPanic() {
	sort.Slice(p.Test.Events, p.Test.Less)

	// delete the last 2 lines:
	// 1. FAIL	github.com/mfridman/tparse/tests	0.012s
	// 2. empty line from summary which has no output
	p.Test.Events = p.Test.Events[:len(p.Test.Events)-2]

	s := fmt.Sprintf("\nPACKAGE: %s", p.Summary.Package)
	n := make([]string, len(s)+1)

	fmt.Printf("%s\n%s\n", s, strings.Join(n, "-"))

	fmt.Printf("%s\t%s\t%s\n%s\n\n",
		Red("PANIC"),
		strconv.FormatFloat(p.Test.Elapsed(), 'f', 2, 64),
		p.Test.Name,
		p.Test.Stack(),
	)
}
