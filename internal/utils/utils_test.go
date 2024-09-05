package utils

import "testing"

func TestFindLongestCommonPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		paths []string
		want  string
	}{
		{
			paths: []string{},
			want:  "",
		},
		{
			paths: []string{
				"github.com/user/project/pkg",
			},
			want: "",
		},
		{
			paths: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/pkg",
				"github.com/user/project/pkg",
			},
			want: "github.com/user/project/pkg",
		},
		{
			paths: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/cmd",
			},
			want: "github.com/user/project/",
		},
		{
			paths: []string{
				"github.com/user/project/pkg",
				"bitbucket.org/user/project/cmd",
			},
			want: "",
		},
		{
			paths: []string{
				"github.com/user/project/pkg",
				"github.com/user/project/cmd",
				"github.com/user/project/cmd/subcmd",
				"github.com/nonuser/project/cmd/subcmd",
			},
			want: "github.com/",
		},
		{
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
			paths: []string{
				"/",
			},
			want: "",
		},
		{
			paths: []string{
				"/",
				"/",
			},
			want: "/",
		},
		{
			paths: []string{
				"/abc",
				"/abc",
			},
			want: "/abc",
		},
		{
			paths: []string{
				"foo/bar/foo",
				"foo/foo/foo",
			},
			want: "foo/",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := FindLongestCommonPrefix(tt.paths)
			if actual != tt.want {
				t.Errorf("want %s, got %s", tt.want, actual)
			}
		})
	}
}
