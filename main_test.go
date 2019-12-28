package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMarkdownFunc(t *testing.T) {
	markdownText := "# title\nthis is normal text"
	tmpl := markdownFunc(markdownText)
	if string(tmpl) != "<h1>title</h1>\n\n<p>this is normal text</p>\n" {
		t.Error("Illegal HTML generate ", tmpl)
	}
}

func TestMarkdownFileFunc(t *testing.T) {
	markdownFile, err := ioutil.TempFile(os.TempDir(), "sample.*.md")
	if err != nil {
		panic(err)
	}
	defer os.Remove(markdownFile.Name())

	_, err = markdownFile.WriteString("# title\nthis is normal text")
	markdownFile.Sync()
	if err != nil {
		t.Error("Error to write markdown file", err)
	}

	tmpl := markdownFileFunc(markdownFile.Name())
	if string(tmpl) != "<h1>title</h1>\n\n<p>this is normal text</p>\n" {
		t.Error("Illegal HTML generate ", tmpl)
	}
}
