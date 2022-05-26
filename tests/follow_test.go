package parsetest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/mfridman/tparse/internal/app"
	"github.com/mfridman/tparse/internal/check"
	"github.com/mfridman/tparse/parse"
)

func TestFollow(t *testing.T) {
	t.Parallel()

	base := filepath.Join("testdata", "follow")

	tt := []struct {
		fileName string
		err      error
		exitCode int
	}{
		// race detected
		{"test_01", nil, 1},
		{"test_02", nil, 0},
		{"test_03", nil, 0},
		{"test_04", nil, 0},
		{"test_05", parse.ErrNotParseable, 1},
		// build failure in one package
		{"test_06", nil, 2},
	}
	for _, tc := range tt {
		t.Run(tc.fileName, func(t *testing.T) {
			intputFile := filepath.Join(base, tc.fileName+".json")
			options := app.Options{
				FileName:           intputFile,
				FollowOutput:       true,
				DisableTableOutput: true,
			}
			var buf bytes.Buffer
			gotExitCode, err := app.Run(&buf, options)
			if err != nil && !errors.Is(err, tc.err) {
				t.Fatal(err)
			}
			check.Number(t, gotExitCode, tc.exitCode)
			goldenFile := filepath.Join(base, tc.fileName+".golden")
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
