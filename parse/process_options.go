package parse

import "io"

type options struct {
	w      io.Writer
	follow bool
	debug  bool
	events chan<- Event
}

type OptionsFunc func(o *options)

func WithFollowOutput(b bool) OptionsFunc {
	return func(o *options) { o.follow = b }
}

func WithWriter(w io.Writer) OptionsFunc {
	return func(o *options) { o.w = w }
}

func WithDebug() OptionsFunc {
	return func(o *options) { o.debug = true }
}

func WithEvents(c chan<- Event) OptionsFunc {
	return func(o *options) { o.events = c }
}
