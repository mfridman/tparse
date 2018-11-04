package parse

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestSingleFailStack(t *testing.T) {

	// disable colors when testing.
	colors = false

	t.Parallel()

	root := "testdata"
	base := filepath.Join(root, "stack")

	tt := []struct {
		name, input, output string
	}{
		{"input01", filepath.Join(base, "input01.json"), filepath.Join(base, "output01.golden")},
		{"input02", filepath.Join(base, "input02.json"), filepath.Join(base, "output02.golden")},
		{"input03", filepath.Join(base, "input03.json"), filepath.Join(base, "output03.golden")},
		{"input04", filepath.Join(base, "input04.json"), filepath.Join(base, "output04.golden")},
	}

	for _, test := range tt {

		t.Run(test.name, func(t *testing.T) {
			by, err := ioutil.ReadFile(test.input)
			if err != nil {
				t.Fatal(err)
			}

			want, err := ioutil.ReadFile(test.output)
			if err != nil {
				t.Fatal(err)
			}

			pkgs, err := Process(bytes.NewReader(by))
			if err != nil {
				t.Fatal(err)
			}

			for _, pkg := range pkgs {
				failed := pkg.TestsByAction(ActionFail)

				// The input file must be composed of a single test. We're just checking the "stack" output
				// for a single test.
				if len(failed) != 1 {
					t.Fatalf("package %s must contain only 1 test case", pkg.Summary.Package)
				}

				for _, c := range failed {
					got := c.Stack()

					if !bytes.Equal([]byte(got), want) {
						t.Error("failed stack comparison")
					}

					// Note, when go test outputs nested failed tests it favors "spaces" instead of "tabs".
					// If you see a failure here, start with a peak at the quoted got and want strings.
					if t.Failed() {
						t.Logf("\ngot plain:\n%s\ngot quoted:\n%+q", got, got)
						t.Logf("\nwant plain:\n%s\nwant quoted:\n%+q", want, want)
					}
				}
			}

		})

	}
}
