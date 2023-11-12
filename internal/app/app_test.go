package app

import (
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestCommonPackagePrefix(t *testing.T) {
	t.Parallel()
	prefix := findCommonPackagePrefix([]*parse.Package{})
	check.Equal(t, "", prefix)
	// The findCommonPackagePrefix function is not stable if the packages are not sorted.
	// https://github.com/mfridman/tparse/issues/102
	prefix = findCommonPackagePrefix([]*parse.Package{
		{Summary: &parse.Event{Package: "github.com/foo/bar/baz/qux"}},
		{Summary: &parse.Event{Package: "github.com/foo/bar/baz"}},
		{Summary: &parse.Event{Package: "github.com/foo/bar/baz/qux/quux"}},
		{Summary: &parse.Event{Package: "github.com/foo/bar/baz/qux/quux/corge"}},
		{Summary: &parse.Event{Package: "github.com/foo/bar/baz/foo"}},
		{Summary: &parse.Event{Package: "github.com/foo/bar/baz/foo/bar"}},
	})
	check.Equal(t, "github.com/foo/bar", prefix)
}
