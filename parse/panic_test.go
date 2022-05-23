package parse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPanic(t *testing.T) {

	t.Parallel()

	// key is the package name, bool reports whether we expect package to be marked as panicked
	type registry map[string]bool

	root := "testdata"
	base := filepath.Join(root, "panic")

	tt := []struct {
		name string
		registry
	}{
		{"input01.json", registry{
			"github.com/mfridman/tparse/parse": true,
		}},
		{"input02.json", registry{
			"github.com/mfridman/tparse/tests": true,
		}},
		{"input03.json", registry{
			"github.com/mfridman/tparse/tests":  true,
			"github.com/mfridman/tparse/ignore": false,
			"github.com/mfridman/tparse/parse":  false,
			"github.com/mfridman/tparse":        false,
		}},
		{"input04.json", registry{
			"github.com/mfridman/tparse/tests":  true,
			"github.com/mfridman/tparse/parse":  false,
			"github.com/mfridman/tparse":        false,
			"github.com/mfridman/tparse/ignore": false,
		}},
		{"input05.json", registry{
			"github.com/mfridman/tparse/tests":  false,
			"github.com/mfridman/tparse/parse":  true,
			"github.com/mfridman/tparse":        false,
			"github.com/mfridman/tparse/ignore": false,
		}},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f, err := os.Open(filepath.Join(base, test.name))
			if err != nil {
				t.Fatal(err)
			}

			summary, err := Process(f)
			if err != nil {
				t.Fatalf("got error %[1]v of type %[1]T, want nil", err)
			}

			for name, pkg := range summary.Packages {
				want, ok := test.registry[name]
				if !ok {
					t.Log("currently registered packages:")
					for k := range test.registry {
						t.Log(k)
					}
					t.Fatalf("got unmapped package name %q. Check input file and record all unique package names in registry", name)
				}

				if pkg.HasPanic != want {
					t.Log("package: ", name)
					t.Logf("summary: %+v", pkg.Summary)
					t.Fatal("got no panic, expecting package to be marked as has panic")
				}

			}
		})

	}
}
