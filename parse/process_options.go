package parse

import (
	"io"
)

type options struct {
	w             io.Writer
	follow        bool
	followVerbose bool
	debug         bool

	progress       bool
	progressOutput io.Writer
}

type OptionsFunc func(o *options)

func WithFollowOutput(b bool) OptionsFunc {
	return func(o *options) { o.follow = b }
}

func WithFollowVersboseOutput(b bool) OptionsFunc {
	return func(o *options) { o.followVerbose = b }
}

func WithWriter(w io.Writer) OptionsFunc {
	return func(o *options) { o.w = w }
}

func WithDebug() OptionsFunc {
	return func(o *options) { o.debug = true }
}

func WithProgress(b bool) OptionsFunc {
	return func(o *options) { o.progress = b }
}

func WithProgressOutput(w io.Writer) OptionsFunc {
	return func(o *options) { o.progressOutput = w }
}
