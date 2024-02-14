# template-parse-recursive
Package for parsing go templates recursively

By default, go's template.ParseGlob does not traverse folders recursively, and uses only filename without folder name as a template name.

This package goes through subfolders recursively and parses the files matching the glob. The templates have the subfolder path in the name, separated by OS-specific separator, as done by path/filepath.

The template names are as relative to the given folder.

## Example

Use like this

```go
package main

import (
    "html/template"
    "os"

    recurparse "github.com/karelbilek/template-parse-recursive"
)

func main() {
    t, err := recurparse.HTMLParse(
        template.New("templates"), 
        "path/to/templates", 
        "*.html",
    )

    if err != nil {
        panic(err)
    }

    templateUnder := t.Lookup("subdir/subdir/template.html")
    templateUnder.Execute(os.Stdout, nil)
}
```

You can also use with embed.FS

```go
package main

import (
    "html/template"
    "os"
    "embed"

    recurparse "github.com/karelbilek/template-parse-recursive"
)

//go:embed html/*
var content embed.FS

func main() {
    t, err := recurparse.HTMLParseFS(
        template.New("templates"),
        content,
        "*.html",
    )

    if err != nil {
        panic(err)
    }

    templateUnder := t.Lookup("html/subdir/subdir/template.html")
    templateUnder.Execute(os.Stdout, nil)
}
```

## Symlinks
The traversal _does_ follow symlinks, and fails when symlinks are errorneous.

It does *not* handle symlink loop in any way and, in such case, will hang forever and run out of memory.

## Windows

On WSL, the package is working the same way as on Unix-like OSes as it's basically Linux.

On "native" windows, the paths are *using forward slashes* - that is, still have `html/subdir/subdir/`, even when Windows is using backward slashes for a directory separator.

On "native" windows, it will try to follow symlinks (note that shortcuts are NOT symlinks and will not be followed). *HOWEVER*, note that Windows don't play well with symlinks - you need to have admin priviledges to create them; and if you git-clone a repo with symlinks, such as this one, the symlinks might be converted to regular files with addresses in them. It's beyond the scope of this package to deal with all this...

For that reason, there are no symlink tests for Windows.

## macOS
On MacOS, it's technically possible to create a file with "/" in its name, and on Unix layer, it's converted to ":". This is the same way this package deals with it; all the "/" are converted to ":".

## License
BSD 3-Clause License

(C) 2022
