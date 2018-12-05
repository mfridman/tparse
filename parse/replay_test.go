package parse

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestReplay(t *testing.T) {

	root := "testdata"
	base := filepath.Join(root, "replay")

	tt := []struct {
		name, input, output string
	}{
		{"input01", filepath.Join(base, "input01.json"), filepath.Join(base, "output01.golden")},
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
			ReplayOutput(&buf, f)

			if !bytes.Equal(buf.Bytes(), want) {
				t.Error("replay input does not match expected output; diff files in replay dir suffixed with .FAIL to debug")
				t.Logf("diff %v %v", "parse/"+test.input+".FAIL", "parse/"+test.output+".FAIL")
				if err := ioutil.WriteFile(test.input+".FAIL", buf.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				if err := ioutil.WriteFile(test.output+".FAIL", want, 0644); err != nil {
					t.Fatal(err)
				}
			}
		})

	}
}
