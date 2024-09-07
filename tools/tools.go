//go:build tools
// +build tools

package tools

// Manage dependencies
//
// https://github.com/golang/go/issues/48429
//
//nolint
import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
