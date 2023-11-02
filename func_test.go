package recurparse

import (
	"embed"
	"html/template"
	"strings"
	"testing"

	_ "embed"
)

//go:embed testdata/func/*.txt
var f embed.FS

func upperString(in string) string {
	return strings.ToUpper(in)
}

func TestFunc(t *testing.T) {
	tmpl, err := HTMLParseFS(
		template.New("templates").Funcs(template.FuncMap{
			"upperString": upperString,
		}),
		f,
		"*.txt",
	)
	if err != nil {
		panic(err)
	}

	b := &strings.Builder{}
	err = tmpl.ExecuteTemplate(b, "testdata/func/func.txt", struct{ Foo string }{Foo: "bar"})
	if err != nil {
		panic(err)
	}
	if b.String() != "BAR" {
		t.Error("not equal BAR")
	}
}
