package utils

import "testing"

func TestFindLongestCommonPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		paths []string
		want  string
	}{
		{
			name:  "empty names",
			paths: []string{},
			want:  "",
		},
		{
			name: "single name",
			paths: []string{
				"github.com/user/project/pkg",
			},
			want: "",
		},
		{
			name: "two identical modules",
			paths: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/pkg",
			},
			want: "github.com/user/project/pkg",
		},
		{
			name: "two modules with common prefix",
			paths: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/cmd",
			},
			want: "github.com/user/project/",
		},
		{
			name: "two different modules",
			paths: []string{
				"github.com/user/project/pkg",
				"bitbucket.org/user/project/cmd",
			},
			want: "",
		},
		{
			name: "two different modules with common prefix",
			paths: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/cmd",
				"github.com/user/project/cmd/subcmd",
				"github.com/nonuser/project/cmd/subcmd",
			},
			want: "github.com/",
		},
		{
			name: "multiple modules with common prefix",
			paths: []string{
				"github.com/foo/bar/baz/qux",
				"github.com/foo/bar/baz",
				"github.com/foo/bar/baz/qux/quux",
				"github.com/foo/bar/baz/qux/quux/corge",
				"github.com/foo/bar/baz/foo",
				"github.com/foo/bar/baz/foo/bar",
			},
			want: "github.com/foo/bar/",
		},
		{
			name: "one slash",
			paths: []string{
				"/",
			},
			want: "",
		},
		{
			name: "two slashes",
			paths: []string{
				"/",
				"/",
			},
			want: "/",
		},
		{
			name: "prefix only with slash",
			paths: []string{
				"/abc",
				"/abc",
			},
			want: "/abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FindLongestCommonPrefix(tt.paths)
			if actual != tt.want {
				t.Errorf("want %s, got %s", tt.want, actual)
			}
		})
	}
}
