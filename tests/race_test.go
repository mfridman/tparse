package parse

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/parse"
)

var raceTestFiles = []struct {
	name string
}{
	{"test_01"},
	{"test_02"},
	{"test_03"},
	// This is a race directly from Test only.
	{"test_04"},
	// This is a race directly from TestMain with other tests that have failed.
	{"test_05"},
	// This is a race directly from TestMain only.
	{"test_06"},
	// This is a race from a Test that calls into a package that has a race condition. (failed assertion)
	{"test_07"},
	// This is a race from a Test that calls into a package that has a race condition. (passed assertion)
	{"test_08"},
}

func TestRaceDetected(t *testing.T) {
	base := filepath.Join("testdata", "race")

	for _, tc := range raceTestFiles {
		t.Run(tc.name, func(t *testing.T) {
			intputFile := filepath.Join(base, tc.name+".json")
			f, err := os.Open(intputFile)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			packages, err := parse.Process(f)
			if err != nil {
				t.Fatal(err)
			}
			_ = packages
		})
	}
}

func TestRaceReplay(t *testing.T) {
	base := filepath.Join("testdata", "race")

	tt := []struct {
		name string
	}{
		{"test_01"},
		{"test_02"},
		{"test_03"},
		// This is a race directly from Test only.
		{"test_04"},
		// This is a race directly from TestMain with other tests that have failed.
		{"test_05"},
		// This is a race directly from TestMain only.
		{"test_06"},
		// This is a race from a Test that calls into a package that has a race condition. (failed assertion)
		{"test_07"},
		// This is a race from a Test that calls into a package that has a race condition. (passed assertion)
		{"test_08"},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			intputFile := filepath.Join(base, test.name+".json")
			options := app.Options{
				FileName:           intputFile,
				FollowOutput:       true,
				DisableTableOutput: true,
			}
			var buf bytes.Buffer
			if err := app.Run(&buf, options); err != nil {
				t.Fatal(err)
			}
			goldenFile := filepath.Join(base, test.name+".golden")
			want, err := ioutil.ReadFile(goldenFile)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf.Bytes(), want) {
				t.Error("input does not match expected output; diff files in follow dir suffixed with .FAIL to debug")
				t.Logf("diff %v %v",
					"tests/"+goldenFile+".FAIL",
					"tests/"+intputFile+".FAIL",
				)
				if err := ioutil.WriteFile(goldenFile+".FAIL", buf.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
