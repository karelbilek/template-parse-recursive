package recurparse

import (
	"fmt"
	templateHtml "html/template"
	templateText "text/template"

	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
)

func TextParse(t *templateText.Template, dirPath, glob string) (*templateText.Template, error) {
	resolved, err := filepath.EvalSymlinks(dirPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot resolve %q (%w)", dirPath, err)
	}

	files, err := getFiles(resolved, "", glob, nil, map[string]bool{})
	if err != nil {
		return nil, err
	}

	// logic copied from src/html/template/helper.go

	if len(files) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("recurparse: no files matched")
	}

	for _, filename := range files {
		fpath := path.Join(dirPath, filename)

		b, err := ioutil.ReadFile(fpath)
		if err != nil {
			return nil, fmt.Errorf("recurparse: cannot read %q: %w", fpath, err)
		}

		s := string(b)

		// this is copied verbatim from go template.. I always found the rewrite logic a bit confusing,
		// but it is what it is. Let's keep the logic.
		if t == nil {
			t = templateText.New(filename)
		}

		var tmpl *templateText.Template

		if filename == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(filename)
		}

		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("recurparse: cannot parse %q: %w", fpath, err)
		}
	}

	return t, nil
}

func HTMLParse(t *templateHtml.Template, dirPath, glob string) (*templateHtml.Template, error) {
	resolved, err := filepath.EvalSymlinks(dirPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot resolve %q (%w)", dirPath, err)
	}

	files, err := getFiles(resolved, "", glob, nil, map[string]bool{})
	if err != nil {
		return nil, err
	}

	// logic copied from src/html/template/helper.go

	if len(files) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("recurparse: no files matched")
	}

	for _, filename := range files {
		fpath := path.Join(dirPath, filename)

		b, err := ioutil.ReadFile(fpath)
		if err != nil {
			return nil, fmt.Errorf("recurparse: cannot read %q: %w", fpath, err)
		}

		s := string(b)

		// this is copied verbatim from go template.. I always found the rewrite logic a bit confusing,
		// but it is what it is. Let's keep the logic.
		if t == nil {
			t = templateHtml.New(filename)
		}

		var tmpl *templateHtml.Template

		if filename == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(filename)
		}

		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("recurparse: cannot parse %q: %w", fpath, err)
		}
	}

	return t, nil
}

// inspired by code in filepath.Glob
func getFiles(resolvedPath, relativePath, glob string, matched []string, visited map[string]bool) ([]string, error) {
	if visited[resolvedPath] {
		return nil, fmt.Errorf("recurparse: found symlink loop on %q (%q)", relativePath, resolvedPath)
	}

	visited[resolvedPath] = true

	fi, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot get file info about %q: %w", resolvedPath, err)
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("recurparse: file %q is not dir", resolvedPath)
	}

	d, err := os.Open(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: error opening dir %q: %w", resolvedPath, err)
	}

	defer d.Close()

	names, _ := d.Readdirnames(-1)
	sort.Strings(names) // sort to get predictable results in tests

	for _, n := range names {
		pathUnder := path.Join(resolvedPath, n)
		relativeUnder := path.Join(relativePath, n)

		resolvedPathUnder, err := filepath.EvalSymlinks(pathUnder)
		if err != nil {
			return nil, fmt.Errorf("recurparse: cannot get evan symlink %q: %w", pathUnder, err)
		}

		fiUnder, err := os.Stat(pathUnder)
		if err != nil {
			return nil, fmt.Errorf("recurparse: cannot get recursive file info about %q: %w", resolvedPath, err)
		}

		if fiUnder.IsDir() {
			// passing a copy of visited, so we allow symlinks that are not recursive
			visitedCopy := make(map[string]bool, len(visited))
			for k := range visited {
				visitedCopy[k] = true
			}

			matched, err = getFiles(resolvedPathUnder, relativeUnder, glob, matched, visitedCopy)
			if err != nil {
				return nil, err
			}
		} else {
			// matching name _before symlink_
			isMatched, err := filepath.Match(glob, n)
			if err != nil {
				return nil, fmt.Errorf("recurpaste: error in matching %q against %q in %q: %w", glob, n, resolvedPath, err)
			}

			if isMatched {
				matched = append(matched, relativeUnder)
			}
		}
	}

	return matched, nil
}
