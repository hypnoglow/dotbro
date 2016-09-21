package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestCopy(t *testing.T) {
	// set up

	src := "/tmp/dotbro/fileuitls/original.txt"
	content := []byte("Some Content")

	if err := os.MkdirAll(path.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(src, content, 0755); err != nil {
		t.Fatal(err)
	}

	// test

	dest := "/tmp/dotbro/fileutils/copy.txt"
	if err := Copy(src, dest); err != nil {
		t.Error(err)
	}

	copyContent, err := ioutil.ReadFile(dest)
	if err != nil {
		t.Error(err)
	}

	if string(copyContent) != string(content) {
		t.Error(err)
	}

	// tear down

	if err := os.Remove(src); err != nil {
		t.Error(err)
	}
}
