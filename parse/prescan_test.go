package parse

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPrescan(t *testing.T) {

	t.Parallel()

	root := "testdata"
	base := filepath.Join(root, "prescan")

	tt := []struct {
		name string
		desc string
		err  error
	}{
		{"input01.txt", "want <nil> err", nil},
		{"input02.txt", "want failure after reading >50 lines of non-parseable events", ErrNotParseable},
		// logic: unparseable event(s), good event(s), at least one event = fail.
		// Once we get a good event, we expect only good events to follow until EOF.
		{"input03.txt", "want failure when stream contains a bad event(s) -> good event(s) -> bad event", ErrNotParseable},
		{"input04.txt", "want failure reading <50 lines of non-parseable events", ErrNotParseable},
	}

	for _, test := range tt {
		test := test

		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			by, err := ioutil.ReadFile(filepath.Join(base, test.name))
			if err != nil {
				t.Fatal(err)
			}

			_, err = Process(bytes.NewReader(by))
			// retrieve original error.
			err = errors.Unwrap(err)

			// dirty, but does the job.
			if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
				t.Fatalf("%s: got err type %T want %T: %v", test.desc, err, test.err, err)
			}

			if err != nil && err != test.err {
				t.Fatalf("got wrong opaque error %q; want ErrNotParseable", err)
			}

		})

	}
}
