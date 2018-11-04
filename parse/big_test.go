package parse

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestBigOutcome(t *testing.T) {

	t.Parallel()

	// key is the package name, action reports the big outcome for the entire package
	type registry map[string]Action

	root := "testdata"
	base := filepath.Join(root, "big")

	tt := []struct {
		name string
		registry
	}{
		{"input01.json", registry{
			"github.com/mfridman/tparse/tests": ActionFail,
		}},
		{"input02.json", registry{
			"github.com/astromail/rover/tests":          ActionFail,
			"github.com/astromail/rover/cmd/roverd":     ActionPass,
			"github.com/astromail/rover/smtp":           ActionPass,
			"github.com/astromail/rover/storage":        ActionPass,
			"github.com/astromail/rover/errors":         ActionPass,
			"github.com/astromail/rover/storage/badger": ActionPass,
			"github.com/astromail/rover":                ActionPass,
		}},
		{"input03.json", registry{
			"fmt": ActionPass,
		}},
		{"input04.json", registry{
			"github.com/astromail/rover/tests": ActionPass,
		}},
		{"input05.json", registry{
			"github.com/astromail/rover/tests": ActionPass,
		}},
		{"input06.json", registry{
			"fmt": ActionPass,
		}},
		{"input07.json", registry{
			"debug/errorcause": ActionPass,
		}},
		{"input08.json", registry{
			"github.com/awesome/pkg": ActionPass,
		}},
	}

	for _, test := range tt {

		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name, len(test.registry))

			by, err := ioutil.ReadFile(filepath.Join(base, test.name))
			if err != nil {
				t.Fatal(err)
			}

			pkgs, err := Process(bytes.NewReader(by))
			if err != nil {
				t.Fatalf("got error %[1]v of type %[1]T, want nil", err)
			}

			if len(pkgs) == 0 {
				t.Fatal("got zero packages, want at least one or more packages")
			}

			for name, pkg := range pkgs {

				want, ok := test.registry[name]
				if !ok {
					t.Log("currently registered packages:")
					for k := range test.registry {
						t.Log(k)
					}
					t.Fatalf("got unmapped package name %q. Check input file and record all unique package names in registry", name)
				}

				if pkg.Summary.Action != want {
					t.Fatalf("failed package summary action: got %q, want %q", pkg.Summary.Action, want)
				}

				if len(pkg.Tests) == 0 && pkg.Summary.Action != ActionPass {
					t.Fatalf("zero test should always return pass: got %q, want %q", pkg.Summary.Action, ActionPass)
				}

				if (pkg.NoTestFiles || pkg.NoTests) && pkg.Summary.Action != ActionPass {
					t.Fatalf("packages marked as [no tests to run] or [no test files] should always return pass: got %q, want %q", pkg.Summary.Action, ActionPass)
				}

				// As a sanity check, iterate over the tests and make sure
				// all tests actually reflect the package outcome.

				switch pkg.Summary.Action {
				case ActionPass:
					// One or more tests must be marked as either pass or skip.
					// A single skipped test will still yield a pass package outcome.
					for _, test := range pkg.Tests {
						switch test.Status() {
						case ActionPass, ActionSkip:
							continue
						default:
							t.Fatalf("all tests, within a package marked passed, should have a status of pass or skip: got %q", test.Status())
						}
					}
				case ActionFail:
					// One or more tests must be marked as failed.
					var failed bool
					for _, test := range pkg.Tests {
						if test.Status() == ActionFail {
							failed = true
							break
						}
					}

					if !failed {
						t.Fatalf("got no failed tests, want one or more tests to be marked as %q", ActionFail)
					}
				default:
					// Catch all, should never get this.
					t.Fatalf("got package summary action %q, want pass or fail", pkg.Summary.Action)
				}

			}
		})

	}
}
