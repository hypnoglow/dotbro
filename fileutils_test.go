package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestCopy(t *testing.T) {
	testCopyPositive(t)
	testCopyNegativeLstat(t)
	testCopyNegativeSymlink(t)
}

func testCopyPositive(t *testing.T) {
	// set up

	src := "/tmp/dotbro/fileutils/original.txt"
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

	if err := os.Remove(dest); err != nil {
		t.Error(err)
	}
}

func testCopyNegativeLstat(t *testing.T) {
	// set up

	src := "/tmp/dotbro/fileutils/original.txt"

	if err := os.MkdirAll(path.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}

	// no read permissions
	if err := ioutil.WriteFile(src, nil, 0333); err != nil {
		t.Fatal(err)
	}

	// test

	dest := "/tmp/dotbro/fileutils/copy.txt"
	err := Copy(dest, dest)
	if err == nil {
		t.Error("No error!")
	}

	// tear down

	if err := os.Remove(src); err != nil {
		t.Error(err)
	}
}

func testCopyNegativeSymlink(t *testing.T) {
	// set up

	original := "/tmp/dotbro/fileutils/original.txt"

	if err := os.MkdirAll(path.Dir(original), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(original, nil, 0755); err != nil {
		t.Fatal(err)
	}

	symlink := "/tmp/dotbro/fileutils/symlink"
	if err := os.Symlink(original, symlink); err != nil {
		t.Fatal(err)
	}

	// test

	dest := "/tmp/dotbro/fileutils/symlink-copy.txt"
	err := Copy(symlink, dest)
	if err == nil {
		t.Error("No error!")
	}

	// tear down

	if err := os.Remove(original); err != nil {
		t.Error(err)
	}

	if err := os.Remove(symlink); err != nil {
		t.Error(err)
	}
}
