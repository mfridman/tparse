package parsetest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestFinalOutcome(t *testing.T) {
	t.Parallel()
	// key is the package name, action reports the final outcome of the package.
	type registry map[string]parse.Action

	base := filepath.Join("testdata", "outcome")

	tt := []struct {
		fileName string
		exitCode int
		registry
	}{
		{"test_01.json", 1, registry{
			"github.com/mfridman/tparse/tests": parse.ActionFail,
		}},
		{"test_02.json", 1, registry{
			"github.com/astromail/rover/tests":          parse.ActionFail,
			"github.com/astromail/rover/cmd/roverd":     parse.ActionPass,
			"github.com/astromail/rover/smtp":           parse.ActionPass,
			"github.com/astromail/rover/storage":        parse.ActionPass,
			"github.com/astromail/rover/errors":         parse.ActionPass,
			"github.com/astromail/rover/storage/badger": parse.ActionPass,
			"github.com/astromail/rover":                parse.ActionPass,
		}},
		{"test_03.json", 0, registry{
			"fmt": parse.ActionPass,
		}},
		{"test_04.json", 0, registry{
			"github.com/astromail/rover/tests": parse.ActionPass,
		}},
		{"test_05.json", 0, registry{
			"github.com/astromail/rover/tests": parse.ActionPass,
		}},
		{"test_06.json", 0, registry{
			"fmt": parse.ActionPass,
		}},
		{"test_07.json", 0, registry{
			"debug/errorcause": parse.ActionPass,
		}},
		{"test_08.json", 0, registry{
			"github.com/awesome/pkg": parse.ActionPass,
		}},
	}
	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			f, err := os.Open(filepath.Join(base, tc.fileName))
			if err != nil {
				t.Fatal(err)
			}
			summary, err := parse.Process(f)
			check.NoError(t, err)
			check.Number(t, len(summary.Packages), len(tc.registry))
			check.Number(t, summary.ExitCode(), tc.exitCode)

			for name, pkg := range summary.Packages {
				want, ok := tc.registry[name]
				if !ok {
					t.Log("currently registered packages:")
					for k := range tc.registry {
						t.Log(k)
					}
					t.Fatalf("got unmapped package name %q. Check input file and record all unique package names in registry", name)
				}
				if pkg.Summary.Action != want {
					t.Fatalf("failed package summary action: got: %q, want: %q", pkg.Summary.Action, want)
				}
				if len(pkg.Tests) == 0 && pkg.Summary.Action != parse.ActionPass {
					t.Fatalf("zero test should always return pass: got: %q, want: %q", pkg.Summary.Action, parse.ActionPass)
				}
				if (pkg.NoTestFiles || pkg.NoTests) && pkg.Summary.Action != parse.ActionPass {
					t.Fatalf("packages marked as [no tests to run] or [no test files] should always return pass: got: %q, want: %q",
						pkg.Summary.Action,
						parse.ActionPass,
					)
				}

				// As a sanity check, iterate over the tests and make sure all tests actually
				// reflect the package outcome.

				switch pkg.Summary.Action {
				case parse.ActionPass:
					// One or more tests must be marked as either pass or skip.
					// A single skipped test will still yield a pass package outcome.
					for _, test := range pkg.Tests {
						switch test.Status() {
						case parse.ActionPass, parse.ActionSkip:
							continue
						default:
							t.Fatalf("all tests within a passed package should have a status of pass or skip: got: %q", test.Status())
						}
					}
				case parse.ActionFail:
					// One or more tests must be marked as failed.
					var failed bool
					for _, tc := range pkg.Tests {
						if tc.Status() == parse.ActionFail {
							failed = true
							break
						}
					}
					if !failed {
						t.Fatalf("got no failed tests, want one or more tests to be marked as: %q", parse.ActionFail)
					}
				default:
					// Catch all, should never get this.
					t.Fatalf("got package summary action %q, want pass or fail", pkg.Summary.Action)
				}
			}
		})
	}
}
