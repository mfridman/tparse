package parse_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/parse"
)

func TestFollow(t *testing.T) {
	base := filepath.Join("testdata", "follow")

	tt := []struct {
		name string
		err  error
	}{
		{"test_01", nil},
		{"test_02", nil},
		{"test_03", nil},
		{"test_04", nil},
		{"test_05", parse.ErrNotParseable},
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
			if err := app.Run(&buf, options); err != nil && !errors.Is(err, test.err) {
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
