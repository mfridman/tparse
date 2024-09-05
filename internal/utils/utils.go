package utils

import (
	"sort"
	"strings"
)

// FindLongestCommonPrefix finds the longest common prefix amongst a list of full package names. For
// example, given the following:
//
//	github.com/owner/repo/cmd/foo
//	github.com/owner/repo/cmd/bar
//
// The longest common prefix is: github.com/owner/repo/cmd
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
		return commonPrefix[:n]
	}
	return ""
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
