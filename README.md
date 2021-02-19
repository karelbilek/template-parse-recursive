# template-parse-recursive
Package for parsing go templates recursively

Note: this might be actually implemented already in go 1.16 https://golang.org/pkg/html/template/#ParseFS . I am not sure :)

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

By default, go's template.ParseGlob does not traverse folders recursively, and uses only filename without folder name as a template name.

This package goes through subfolders recursively and parses the files matching the glob. The templates have the subfolder path in the name, separated by OS-specific separator, as done by path/filepath.

The template names are as relative to the given folder.

It _does_ follow symlinks, and fails when symlinks are errorneous. 

In case of symlink loop, returns error to prevent infinite recursion.

The package was not tested at Windows as I don't own a Windows machine currently; you are free to fork and/or test and/or send PRs.
