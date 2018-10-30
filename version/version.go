//go:generate go run makeversion.go

// Package version is used by the release process to add a git tag, if available.
package version

// The tag is overwritten by the init function in makeversion.go.
var (
	GitTag = ""
)

// Version returns a newline-terminated string describing the current
// version of the build.
func Version() string {
	if GitTag == "" {
		return "devel"
	}
	return GitTag
}
