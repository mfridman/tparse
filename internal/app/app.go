package app

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/mfridman/tparse/parse"
)

type Options struct {
	DisableTableOutput bool
	FollowOutput       bool
	DisableColor       bool
	Format             OutputFormat
	ShowNoTests        bool
	FileName           string

	// TODO(mf): implement
	Progress bool
}

func Run(w io.Writer, option Options) error {
	var reader io.ReadCloser
	var err error
	if option.FileName != "" {
		if reader, err = os.Open(option.FileName); err != nil {
			return err
		}
	} else {
		if reader, err = newPipeReader(); err != nil {
			return errors.New("stdin must be a pipe, or use -file to open go test output file")
		}
	}
	defer reader.Close()

	packages, err := parse.Process(
		reader,
		parse.WithFollowOutput(option.FollowOutput),
		parse.WithWriter(w),
	)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		return fmt.Errorf("found no go test packages")
	}
	// Useful for tests when we don't need additional output.
	if option.DisableTableOutput {
		return nil
	}
	return display(w, packages, option)
}

func newPipeReader() (io.ReadCloser, error) {
	finfo, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	// Check file mode bits to test for named pipe as stdin.
	if finfo.Mode()&os.ModeNamedPipe != 0 {
		return os.Stdin, nil
	}
	return nil, errors.New("stdin must be a pipe")
}
