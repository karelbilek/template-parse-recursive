package recurparse

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMatchingNames(t *testing.T) {
	type matchingNamesDatum struct {
		Directory   string
		Glob        string
		Expected    []string
		ExpectedErr string
	}

	matchingNamesData := []matchingNamesDatum{
		{
			Directory: "testdata/getFiles/1_simple_flat",
			Glob:      "*.html",
			Expected: []string{
				"1.html",
				"3.html",
			},
		},
		{
			Directory: "testdata/getFiles/2_simple_dirs",
			Glob:      "*.html",
			Expected: []string{
				"1.html",
				"3.html",
				"first/4.html",
				"second/7.html",
			},
		},
	}
	if runtime.GOOS != "windows" {
		matchingNamesData = append(matchingNamesData, matchingNamesDatum{
			Directory: "testdata/getFiles/3_unix_symlink",
			Glob:      "*.html",
			Expected: []string{
				"1.html",
				"3.html",
				"first/4.html",
				"second/4.html",
			},
		})
	}

TEST:
	for i, d := range matchingNamesData {
		resolved, err := filepath.EvalSymlinks(d.Directory)
		if err != nil {
			t.Fatalf("test %d: cannot resolve %q: %+v", i, d.Directory, err)
		}

		fsys := os.DirFS(resolved)

		files, err := matchingNames(fsys, d.Glob)

		if d.ExpectedErr != "" {
			if err == nil || err.Error() != d.ExpectedErr {
				t.Fatalf("Expected error %q, got %+v", d.ExpectedErr, err)
			}
			continue TEST
		}

		if err != nil {
			t.Fatalf("test %d (%s): err %+v", i, d.Directory, err)
		}

		if len(files) != len(d.Expected) {
			for _, f := range files {
				fmt.Println(f)
			}
			t.Fatalf("test %d (%s): different lengths: %d vs %d", i, d.Directory, len(files), len(d.Expected))
		}

		for j := range files {
			if files[j] != d.Expected[j] {
				t.Fatalf("test %d (%s): %d : %q != %q", i, d.Directory, j, files[j], d.Expected[j])
			}
		}
	}
}
