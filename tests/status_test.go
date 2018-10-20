package parse_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mfridman/tparse/parse"
)

func TestStatus(t *testing.T) {

	t.Parallel()

	filepath.Walk("./testdata/big", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() && filepath.Ext(info.Name()) != ".json" {
			return nil
		}

		actionsByFilename := []struct {
			prefix string
			parse.Action
		}{
			{"pass", parse.ActionPass},
			// although one or more test(s) are marked as skipped, the test is considered passed.
			{"skip", parse.ActionPass},
			{"fail", parse.ActionFail},
		}

		t.Run(info.Name(), func(t *testing.T) {

			var want parse.Action
			var supported bool

			for i := range actionsByFilename {
				if strings.HasPrefix(info.Name(), actionsByFilename[i].prefix) {
					want = actionsByFilename[i].Action
					supported = true
					break
				}
			}

			if !supported {
				t.Fatalf("got unsupported filename %q; want testdata/big file name prefixed pass|fail|skip", info.Name())
			}

			by, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			pkgs, err := parse.Do(bytes.NewReader(by))
			if err != nil {
				t.Fatalf("failed to parse event: %v", err)
			}

			for _, pkg := range pkgs {
				if pkg.Summary.Action != want {
					t.Logf("log: file: %s", info.Name())
					t.Fatalf("failed package summary action: got %q, want %q", pkg.Summary.Action, want)
				}

				// zero tests [no tests to run] should always yield a pass
				if len(pkg.Tests) == 0 && pkg.Summary.Action != parse.ActionPass {
					t.Fatalf("zero test should always return pass: got %q, want %q", pkg.Summary.Action, parse.ActionPass)
				}

				// As a sanity check, we're going to iterate over the tests and make sure
				// it reflects the correct package outcome

				switch pkg.Summary.Action {
				case parse.ActionPass:
					// one or more tests must be explicitly marked as either pass or skip
					// anything else should result in the test failing
					for _, test := range pkg.Tests {
						switch test.Status() {
						case parse.ActionPass, parse.ActionSkip:
							continue
						default:
							t.Fatalf("all tests, within a package marked passed, should have a status of pass or skip: got %q", test.Status())
						}
					}
				case parse.ActionFail:
					// one or more tests must be marked as failed, otherwise the test is considered failed
					var failed bool
					for _, test := range pkg.Tests {
						if test.Status() == parse.ActionFail {
							failed = true
							break
						}
					}

					if !failed {
						t.Fatalf("got no failed tests, want one or more tests to be marked as %q", parse.ActionFail)
					}
				default:
					// catch all failure, should never occur
					t.Fatalf("got package summary action %q, want pass or fail", pkg.Summary.Action)
				}

			}
		})

		return nil
	})

}
