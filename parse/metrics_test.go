package parse

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestMetrics(t *testing.T) {

	t.Parallel()

	// This test depends on metrics_test.json, which contains the output of 9 std lib packages:
	// go test -count=1 fmt strings bytes bufio crypto log mime sort time -json

	expected := []string{"fmt", "strings", "bytes", "bufio", "crypto", "log", "mime", "sort", "time"}

	f := "./testdata/metrics_test.json"
	by, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}

	pkgs, err := Start(bytes.NewReader(by))
	if err != nil {
		t.Fatal(err)
	}

	if len(pkgs) != 9 {
		t.Logf("file: %s", f)
		t.Fatalf("got %d packages, want 9 packages (known ahead of time)", len(pkgs))
	}

	m := map[string]bool{}
	for _, s := range expected {
		m[s] = true
	}

	for name, pkg := range pkgs {
		if _, ok := m[name]; !ok {
			t.Errorf("got unknown packages %q, want one of:\n%s", name, strings.Join(expected, ", "))
		}

		if pkg.Summary == nil {
			t.Fatalf("package %q cannot contain nil summary", name)
		}

		if pkg.Summary.Action != ActionPass {
			t.Errorf("unexpected action %v, want %v", pkg.Summary.Action, ActionPass)
		}
	}

	if t.Failed() {
		t.FailNow()
	}

	tests := []struct {
		name                           string
		total, passed, skipped, failed int
		elapsed                        float64
	}{
		{"fmt", 59, 58, 1, 0, 0.22},
		{"strings", 107, 107, 0, 0, 5.494},
		{"bytes", 123, 123, 0, 0, 3.5380000000000003},
		{"bufio", 69, 69, 0, 0, 0.07},
		{"crypto", 5, 5, 0, 0, 0.016},
		{"log", 8, 8, 0, 0, 0.085},
		{"mime", 20, 20, 0, 0, 0.025},
		{"sort", 37, 36, 1, 0, 3.117},
		{"time", 118, 117, 1, 0, 7.157},
	}

	for _, test := range tests {
		t.Run(test.name+"_test", func(t *testing.T) {
			pkg := pkgs[test.name]

			if len(pkg.Tests) != test.total {
				t.Fatalf("got %d total tests in package %q, want %d total tests", len(pkg.Tests), test.name, test.total)
			}

			pa := pkg.TestsByAction(ActionPass)
			if len(pa) != test.passed {
				t.Fatalf("got %d passed tests in package %q, want %d passed tests", len(pa), test.name, test.passed)
			}

			sk := pkg.TestsByAction(ActionSkip)
			if len(sk) != test.skipped {
				t.Fatalf("got %d passed tests in package %q, want %d passed tests", len(sk), test.name, test.skipped)
			}

			fa := pkg.TestsByAction(ActionFail)
			if len(fa) != test.failed {
				t.Fatalf("got %d failed tests in package %q, want %d failed tests", len(fa), test.name, test.failed)
			}

			if pkg.Summary.Elapsed != test.elapsed {
				t.Fatalf("got elapsed time %f for package %q, want %f", pkg.Summary.Elapsed, test.name, test.elapsed)
			}
		})
	}
}
