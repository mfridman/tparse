package parse

import (
	"fmt"
)

type PanicErr struct {
	*Package
}

var _ error = (*PanicErr)(nil)

func (p *PanicErr) Error() string {
	return fmt.Sprintf("panic: package: %+v", p.Package.Summary)
}
