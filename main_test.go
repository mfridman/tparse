package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunExitCode(t *testing.T) {
	t.Parallel()

	passFile := filepath.Join("tests", "testdata", "outcome", "test_03.jsonl")
	failFile := filepath.Join("tests", "testdata", "outcome", "test_01.jsonl")
	missingLogFile := filepath.Join(t.TempDir(), "missing", "test.log")

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string
	}{
		{
			name:     "all tests pass",
			args:     []string{"-file", passFile},
			wantCode: 0,
		},
		{
			name:     "test failure",
			args:     []string{"-file", failFile},
			wantCode: 1,
		},
		{
			name:       "unknown flag",
			args:       []string{"-file", passFile, "-bogus"},
			wantCode:   2,
			wantStderr: "flag provided but not defined: -bogus",
		},
		{
			name:       "invalid format",
			args:       []string{"-file", passFile, "-format", "bogus"},
			wantCode:   2,
			wantStderr: `invalid option:"bogus". The -format flag must be one of`,
		},
		{
			name:       "invalid sort",
			args:       []string{"-file", passFile, "-sort", "bogus"},
			wantCode:   2,
			wantStderr: `invalid option:"bogus". The -sort flag must be one of`,
		},
		{
			name:       "follow output file cannot be created",
			args:       []string{"-file", passFile, "-follow-output", missingLogFile},
			wantCode:   1,
			wantStderr: "no such file or directory",
		},
		{
			name:       "missing input file",
			args:       []string{"-file", filepath.Join(t.TempDir(), "missing.jsonl")},
			wantCode:   1,
			wantStderr: "no such file or directory",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			gotCode := run(append(tc.args, "-nocolor"), &stdout, &stderr)
			assert.Equal(t, tc.wantCode, gotCode, "stderr: %s", stderr.String())
			if tc.wantStderr != "" {
				assert.Contains(t, stderr.String(), tc.wantStderr)
			}
		})
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()

	for _, arg := range []string{"-h", "-help"} {
		t.Run(arg, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			gotCode := run([]string{arg}, &stdout, &stderr)
			assert.Equal(t, 0, gotCode)
			assert.Empty(t, stderr.String())
			assert.Equal(t, usage, stdout.String())
		})
	}
}

func TestUsageLineLength(t *testing.T) {
	t.Parallel()

	for i, line := range strings.Split(usage, "\n") {
		assert.LessOrEqual(t, len(line), 80, "usage line %d is too long: %q", i+1, line)
		assert.Equal(t, strings.TrimRight(line, " \t"), line, "usage line %d has trailing space", i+1)
	}
}
