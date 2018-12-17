package parse

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRaceReplay(t *testing.T) {

	root := "testdata"
	base := filepath.Join(root, "race")

	tt := []struct {
		name, input, output string
	}{
		{"input01", filepath.Join(base, "input01.json"), filepath.Join(base, "output01.golden")},
		{"input02", filepath.Join(base, "input02.json"), filepath.Join(base, "output02.golden")},
		{"input03", filepath.Join(base, "input03.json"), filepath.Join(base, "output03.golden")},
	}

	for _, test := range tt {

		t.Run(test.name, func(t *testing.T) {
			want, err := ioutil.ReadFile(test.output)
			if err != nil {
				t.Fatal(err)
			}

			f, err := os.Open(test.input)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			ReplayRaceOutput(&buf, f)

			if !bytes.Equal(buf.Bytes(), want) {
				t.Error("race input does not match expected output; diff files in race dir suffixed with .FAIL to debug")
				t.Logf("diff %v %v", "parse/"+test.input+".FAIL", "parse/"+test.output+".FAIL")
				if err := ioutil.WriteFile(test.input+".FAIL", buf.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				if err := ioutil.WriteFile(test.output+".FAIL", want, 0644); err != nil {
					t.Fatal(err)
				}
			}

			if _, err := f.Seek(0, 0); err != nil {
				t.Fatalf("failed to seek back to begining of file: %v", err)
			}

			if _, err := Process(f); err != ErrRaceDetected {
				t.Fatalf("got wrong opaque error %q; want ErrRaceDetected", err)
			}

		})

	}
}
