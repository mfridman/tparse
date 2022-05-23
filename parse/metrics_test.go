package parse

import (
	"os"
	"strings"
	"testing"
)

func TestMetrics(t *testing.T) {

	t.Parallel()

	// This test depends on metrics_test.json, which contains the output of 9 std lib packages:
	// go test -count=1 fmt strings bytes bufio crypto log mime sort time -json

	expected := []string{"fmt", "strings", "bytes", "bufio", "crypto", "log", "mime", "sort", "time"}

	f, err := os.Open("./testdata/metrics_test.json")
	if err != nil {
		t.Fatal(err)
	}
	summary, err := Process(f)
	if err != nil {
		t.Fatal(err)
	}

	if l := len(summary.Packages); l != 9 {
		t.Logf("file: %s", f)
		t.Fatalf("got %d packages, want 9 packages (known ahead of time)", l)
	}

	m := map[string]bool{}
	for _, s := range expected {
		m[s] = true
	}

	for name, pkg := range summary.Packages {
		if _, ok := m[name]; !ok {
			t.Errorf("got unknown packages %q, want one of:\n%s", name, strings.Join(expected, ", "))
		}

		if pkg.Summary == nil {
			t.Fatalf("package %q cannot contain nil summary", name)
		}

		if pkg.Summary.Action != ActionPass {
			t.Logf("failed pkg: %v", name)
			t.Errorf("unexpected action %q, want %q", pkg.Summary.Action, ActionPass)
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
			pkg := summary.Packages[test.name]

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

func TestElapsed(t *testing.T) {

	t.Parallel()

	// This test depends on elapsed_test.json, which contains the output of 2 std lib tests
	// with known elapsed time.
	// go test -count=1 strings -run="^(TestCompareStrings|TestCaseConsistency$)" -json -cover

	expected := map[string]float64{
		"TestCompareStrings":  3.49,
		"TestCaseConsistency": 0.17,
	}

	fileName := "./testdata/elapsed_test.json"
	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := Process(f)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(summary.Packages); l != 1 {
		for n := range summary.Packages {
			t.Log("got pkg name:", n)
		}
		t.Fatalf("got %d packages, want one package: strings", l)
	}

	pkg, ok := summary.Packages["strings"]
	if !ok {
		t.Fatalf("got unexpected pkg: %v\nwant strings", pkg)
	}

	if len(pkg.Tests) != 2 {
		t.Fatalf("got %d tests, want two", len(pkg.Tests))
	}

	for _, test := range pkg.Tests {
		wantElapsed, ok := expected[test.Name]
		if !ok {
			t.Errorf("got unknown test name %q", test.Name)
		}
		if test.Elapsed() != wantElapsed {
			t.Errorf("got %v elapsed time for test: %q, want %v", test.Elapsed(), test.Name, wantElapsed)
		}
	}

	if t.Failed() {
		t.Log("expected test names:")
		for name := range expected {
			t.Log(name)
		}
	}
}
