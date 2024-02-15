package recurparse

import (
	"html/template"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func ExampleHTMLParseFS() {
	fsTest := fstest.MapFS{
		"data.txt": &fstest.MapFile{
			Data: []byte("super simple {{.Foo}}"),
		},
		"some/deep/file.txt": &fstest.MapFile{
			Data: []byte("other simple {{.Foo}}"),
		},
	}

	tmpl, err := HTMLParseFS(
		nil,
		fsTest,
		"*.txt",
	)
	if err != nil {
		panic(err)
	}

	err = tmpl.ExecuteTemplate(os.Stdout, "some/deep/file.txt", struct{ Foo string }{Foo: "bar"})
	if err != nil {
		panic(err)
	}

	// Output: other simple bar
}

func TestSimpleExisting(t *testing.T) {
	fsTest := fstest.MapFS{
		"data.txt": &fstest.MapFile{
			Data: []byte("super simple {{.Foo}}"),
		},
		"some/deep/file.txt": &fstest.MapFile{
			Data: []byte("other simple {{.Foo}}"),
		},
	}

	existing, err := template.New("existing").Parse("existing {{.Foo}}")
	if err != nil {
		panic(err)
	}

	tmpl, err := HTMLParseFS(
		existing,
		fsTest,
		"*.txt",
	)
	if err != nil {
		panic(err)
	}

	b := &strings.Builder{}
	err = tmpl.ExecuteTemplate(b, "some/deep/file.txt", struct{ Foo string }{Foo: "bar"})
	if err != nil {
		panic(err)
	}
	if b.String() != "other simple bar" {
		t.Error("not equal other simple bar")
	}

	b = &strings.Builder{}
	err = tmpl.ExecuteTemplate(b, "existing", struct{ Foo string }{Foo: "bar"})
	if err != nil {
		panic(err)
	}

	if b.String() != "existing bar" {
		t.Error("not equal super simple bar")
	}
}
