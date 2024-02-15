// Package recurparse parsing go templates recursively, instead of default template behavior
// that puts all files together.
//
// It goes through subfolders recursively and parses the files matching the glob.
// The templates have the subfolder path in the name, separated by forward slash (even on windows).
//
// The template names are as relative to the given folder.
//
// All the 4 functions behave in similar way.
//
// If the first argument is nil,
// the resulting template will have one of the files as name and content;
// if it's an existing template, it will add the files as associated templates.
//
// The pattern works only on the final filename; that is, k*.html will match foo/bar/kxxx.html;
// it does NOT filter the directory name, all directories are walked through.
//
// The matching logic is using filepath.Match on the filename, in the same way template.Parse do it.
// It follows all symlinks, the symlinks will be there under the symlink name.
// If there is a "symlink loop" (that is, symlink to .. or similar), the function will panic and run out of memory.
//
// If there is no files that matches, the function errors, same as go's ParseFiles.
package recurparse

import (
	"fmt"
	templateHtml "html/template"
	templateText "text/template"

	"io/fs"
	"os"
	"path/filepath"
)

// named templ, as `template` is import name in tests
// this is all we actually need for template type
type templ interface {
	comparable
	Name() string
}

// this is abstraction of top-level template functions, so I can reuse same functions
type templateCreator[T templ] interface {
	New(name string) T
	NewBasedOn(nameGiver T, name string) T
	Parse(nameGiver T, text string) (T, error)
	Nil() T
}

type htmlTemplateCreator struct{}

func (htmlTemplateCreator) New(name string) *templateHtml.Template {
	return templateHtml.New(name)
}

func (htmlTemplateCreator) NewBasedOn(t *templateHtml.Template, name string) *templateHtml.Template {
	return t.New(name)
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

func (textTemplateCreator) NewBasedOn(t *templateText.Template, name string) *templateText.Template {
	return t.New(name)
}

func (textTemplateCreator) Parse(t *templateText.Template, text string) (*templateText.Template, error) {
	return t.Parse(text)
}

func (textTemplateCreator) Nil() *templateText.Template {
	return nil
}

func parseFS[T templ](t T, creator templateCreator[T], fsys fs.FS, glob string) (T, error) {
	// first we get the names
	n := creator.Nil()
	files, err := matchingNames(fsys, glob)
	if err != nil {
		return n, err
	}

	if len(files) == 0 {
		// Not really a problem, but be consistent with ParseFiles
		return n, fmt.Errorf("recurparse: no files matched")
	}

	// now parse the templates.
	// the actual code just logic copied from src/html/template/helper.go, just changed for our purposes

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
			tmpl = creator.NewBasedOn(t, filename)
		}

		_, err = creator.Parse(tmpl, s)
		if err != nil {
			return n, fmt.Errorf("recurparse: cannot parse %q: %w", filename, err)
		}
	}

	return t, nil
}

// TextParseFS opens a fs.FS filesystem and recursively parses the files there as text templates.
//
// See package docs for details of the behavior.
func TextParseFS(t *templateText.Template, fsys fs.FS, glob string) (*templateText.Template, error) {
	return parseFS[*templateText.Template](t, textTemplateCreator{}, fsys, glob)
}

// TextParse opens a directory and recursively parses the files there as text templates.
//
// See package docs for details of the behavior.
func TextParse(t *templateText.Template, dirPath, glob string) (*templateText.Template, error) {
	resolved, err := filepath.EvalSymlinks(dirPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot resolve %q (%w)", dirPath, err)
	}

	fsys := os.DirFS(resolved)
	return TextParseFS(t, fsys, glob)
}

// HTMLParseFS opens a fs.FS filesystem and recursively parses the files there as HTML templates.
//
// See package docs for details of the behavior.
func HTMLParseFS(t *templateHtml.Template, fsys fs.FS, glob string) (*templateHtml.Template, error) {
	return parseFS[*templateHtml.Template](t, htmlTemplateCreator{}, fsys, glob)
}

// HTMLParse opens a fs.FS filesystem and recursively parses the files there as HTML templates.
//
// See package docs for details of the behavior.
func HTMLParse(t *templateHtml.Template, dirPath, glob string) (*templateHtml.Template, error) {
	resolved, err := filepath.EvalSymlinks(dirPath)
	if err != nil {
		return nil, fmt.Errorf("recurparse: cannot resolve %q (%w)", dirPath, err)
	}

	fsys := os.DirFS(resolved)
	return HTMLParseFS(t, fsys, glob)
}

// matchingNames is where we walk through the FS and actually get the names
func matchingNames(myfs fs.FS, glob string) ([]string, error) {
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
			// we return errors everywhere we can... not sure if it's the best idea,
			// but I guess better error than not? be safe
			if err != nil {
				return err
			}
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
							if err != nil {
								return err
							}
							isDir := fsStat.IsDir()
							if isDir {
								err := walk(path)
								if err != nil {
									return err
								}
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
