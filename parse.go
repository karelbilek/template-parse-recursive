package recurparse

import (
	"fmt"
	templateHtml "html/template"
	templateText "text/template"

	"io/fs"
	"os"
	"path/filepath"
)

type nameGiver interface {
	comparable
	Name() string
}

type templateCreator[T nameGiver] interface {
	New(name string) T
	Parse(nameGiver T, text string) (T, error)
	Nil() T
}

type htmlTemplateCreator struct{}

func (htmlTemplateCreator) New(name string) *templateHtml.Template {
	return templateHtml.New(name)
}

func (htmlTemplateCreator) Parse(t *templateHtml.Template, text string) (*templateHtml.Template, error) {
	return t.Parse(text)
}

func (htmlTemplateCreator) Nil() *templateHtml.Template {
	return nil
}

type textTemplateCreator struct{}

func (textTemplateCreator) New(name string) *templateText.Template {
	return templateText.New(name)
}

func (textTemplateCreator) Parse(t *templateText.Template, text string) (*templateText.Template, error) {
	return t.Parse(text)
}

func (textTemplateCreator) Nil() *templateText.Template {
	return nil
}

func parseFS[T nameGiver](t T, creator templateCreator[T], fsys fs.FS, glob string) (T, error) {
	n := creator.Nil()
	files, err := getFilesFS(fsys, glob)
	if err != nil {
		return n, err
	}
	// logic copied from src/html/template/helper.go

	if len(files) == 0 {
		// Not really a problem, but be consistent.
		return n, fmt.Errorf("recurparse: no files matched")
	}

	for _, filename := range files {
		b, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return n, fmt.Errorf("recurparse: cannot read %q: %w", filename, err)
		}

		s := string(b)

		// this is copied verbatim from go template.. I always found the rewrite logic a bit confusing,
		// but it is what it is. Let's keep the logic.
		if t == n {
			t = creator.New(filename)
		}

		var tmpl T

		if filename == t.Name() {
			tmpl = t
		} else {
			tmpl = creator.New(filename)
		}

		_, err = creator.Parse(tmpl, s)
		if err != nil {
			return n, fmt.Errorf("recurparse: cannot parse %q: %w", filename, err)
		}
	}

	return t, nil
}

func TextParseFS(t *templateText.Template, fsys fs.FS, glob string) (*templateText.Template, error) {
	return parseFS[*templateText.Template](t, textTemplateCreator{}, fsys, glob)
}

func TextParse(t *templateText.Template, dirPath, glob string) (*templateText.Template, error) {
	resolved, err := filepath.EvalSymlinks(dirPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot resolve %q (%w)", dirPath, err)
	}

	fsys := os.DirFS(resolved)
	return TextParseFS(t, fsys, glob)
}

func HTMLParseFS(t *templateHtml.Template, fsys fs.FS, glob string) (*templateHtml.Template, error) {
	return parseFS[*templateHtml.Template](t, htmlTemplateCreator{}, fsys, glob)
}

func HTMLParse(t *templateHtml.Template, dirPath, glob string) (*templateHtml.Template, error) {
	resolved, err := filepath.EvalSymlinks(dirPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot resolve %q (%w)", dirPath, err)
	}

	fsys := os.DirFS(resolved)
	return HTMLParseFS(t, fsys, glob)
}

func getFilesFS(myfs fs.FS, glob string) ([]string, error) {
	isSymlink := func(d fs.DirEntry) (bool, error) {
		info, err := d.Info()
		if err != nil {
			return false, err
		}

		return info.Mode()&os.ModeSymlink != 0, nil
	}

	var matched []string

	var walk func(dir string) error
	walk = func(dir string) error {
		err := fs.WalkDir(myfs, dir, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				isMatched, err := filepath.Match(glob, d.Name())
				if err != nil {
					return err
				}
				if isMatched {
					matched = append(matched, path)
				} else {
					sym, err := isSymlink(d)
					if err != nil {
						return err
					}
					if sym {
						fsStat, err := fs.Stat(myfs, path)
						if err == nil {
							// ignore error here; cen be non-existent symlink target
							if err != nil {
								return err
							}
							isDir := fsStat.IsDir()
							if isDir {
								walk(path)
							}
						}
					}
				}
			}
			return nil
		})
		return err
	}

	err := walk(".")

	return matched, err
}
