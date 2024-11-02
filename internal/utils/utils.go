package utils

import (
	"io"
	"sort"
	"strings"
)

// FindLongestCommonPrefix finds the longest common path prefix of a set of paths. For example,
// given the following:
//
//	github.com/owner/repo/cmd/foo
//	github.com/owner/repo/cmd/bar
//
// The longest common prefix is: github.com/owner/repo/cmd/ (note the trailing slash is included).
func FindLongestCommonPrefix(paths []string) string {
	if len(paths) < 2 {
		return ""
	}
	// Sort the paths to optimize comparison.
	sort.Strings(paths)

	first, last := paths[0], paths[len(paths)-1]
	if first == last {
		return first
	}

	// Find the common prefix between the first and last sorted paths.
	commonPrefixLength := 0
	minLength := minimum(len(first), len(last))
	for commonPrefixLength < minLength && first[commonPrefixLength] == last[commonPrefixLength] {
		commonPrefixLength++
	}

	// Ensure the common prefix ends at a boundary.
	commonPrefix := first[:commonPrefixLength]
	if n := strings.LastIndex(commonPrefix, "/"); n != -1 {
		return commonPrefix[:n+1]
	}
	return ""
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DiscardCloser is an io.Writer that implements io.Closer by doing nothing.
//
// https://github.com/golang/go/issues/22823
type WriteNopCloser struct {
	io.Writer
}

func (WriteNopCloser) Close() error {
	return nil
}
