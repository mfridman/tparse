package utils

import "testing"

func TestFindLongestCommonPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		names []string
		want  string
	}{
		{
			name:  "empty names",
			names: []string{},
			want:  "",
		},
		{
			name: "single name",
			names: []string{
				"github.com/user/project/pkg",
			},
			want: "",
		},
		{
			name: "two identical modules",
			names: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/pkg",
			},
			want: "github.com/user/project/pkg",
		},
		{
			name: "two modules with common prefix",
			names: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/cmd",
			},
			want: "github.com/user/project",
		},
		{
			name: "two different modules",
			names: []string{
				"github.com/user/project/pkg",
				"bitbucket.org/user/project/cmd",
			},
			want: "",
		},
		{
			name: "two different modules with common prefix",
			names: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/cmd",
				"github.com/user/project/cmd/subcmd",
				"github.com/nonuser/project/cmd/subcmd",
			},
			want: "github.com",
		},
		{
			name: "multiple modules with common prefix",
			names: []string{
				"github.com/foo/bar/baz/qux",
				"github.com/foo/bar/baz",
				"github.com/foo/bar/baz/qux/quux",
				"github.com/foo/bar/baz/qux/quux/corge",
				"github.com/foo/bar/baz/foo",
				"github.com/foo/bar/baz/foo/bar",
			},
			want: "github.com/foo/bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FindLongestCommonPrefix(tt.names)
			if actual != tt.want {
				t.Errorf("want %s, got %s", tt.want, actual)
			}
		})
	}
}
