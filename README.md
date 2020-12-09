# template-parse-recursive
Package for parsing go templates recursively

By default, go's template.ParseGlob only parses the current folder.

This package goes through all subfolders and parses the files matching the glob. The templates have the subfolder path in the name, separated by OS-specific separator, as done by path/filepath.

It _does_ follow symlinks, and fails when symlinks are errorneous.

As it does follow symlink, it's possible to make it run forever with symlink loop. This package makes no attempt to prevent that.

The package was not tested at Windows as I don't own a Windows machine currently; you are free to fork and test.
