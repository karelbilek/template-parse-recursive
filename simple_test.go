package recurparse

import (
	"embed"
	"html/template"
	"strings"
	"testing"
)

//go:embed testdata/simple/*.txt
var simple embed.FS

func TestSimpleNil(t *testing.T) {
	tmpl, err := HTMLParseFS(
		nil,
		simple,
		"*.txt",
	)
	if err != nil {
		panic(err)
	}

	b := &strings.Builder{}
	err = tmpl.ExecuteTemplate(b, "testdata/simple/simple.txt", struct{ Foo string }{Foo: "bar"})
	if err != nil {
		panic(err)
	}
	if b.String() != "super simple bar" {
		t.Error("not equal super simple bar")
	}

}

func TestSimpleExisting(t *testing.T) {
	existing, err := template.New("existing").Parse("existing {{.Foo}}")
	if err != nil {
		panic(err)
	}

	tmpl, err := HTMLParseFS(
		existing,
		simple,
		"*.txt",
	)
	if err != nil {
		panic(err)
	}

	b := &strings.Builder{}
	err = tmpl.ExecuteTemplate(b, "testdata/simple/simple.txt", struct{ Foo string }{Foo: "bar"})
	if err != nil {
		panic(err)
	}
	if b.String() != "super simple bar" {
		t.Error("not equal super simple bar")
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
