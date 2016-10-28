package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestIsExists(t *testing.T) {
	testIsExistsPositive(t)
	testIsExistsNegative(t)
	testIsExistsError(t)
}

func testIsExistsPositive(t *testing.T) {
	// set up

	testPath := "/tmp/dotbro/helpers/file.txt"
	content := []byte("Some Content")

	if err := os.MkdirAll(path.Dir(testPath), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(testPath, content, 0755); err != nil {
		t.Fatal(err)
	}

	// test

	exists, err := isExists(testPath)
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Errorf("Path %s should exist", testPath)
	}

	// tear down

	if err := os.Remove(testPath); err != nil {
		t.Error(err)
	}
}

func testIsExistsNegative(t *testing.T) {
	// set up

	testPath := "/tmp/dotbro/helpers/file.txt"

	_, err := os.Stat(testPath)
	if err == nil {
		// path exists
		if err = os.Remove(testPath); err != nil {
			t.Error(err)
		}
	}

	// test

	exists, err := isExists(testPath)
	if err != nil {
		t.Error(err)
	}

	if exists {
		t.Errorf("Path %s should not exist", testPath)
	}

	// tear down
}

func testIsExistsError(t *testing.T) {
	// set up

	testPath := "/tmp/dotbro/helpers/denied/denied.txt"
	content := []byte("Some Content")

	if err := os.MkdirAll(path.Dir(testPath), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(testPath, content, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(path.Dir(testPath), 0000); err != nil {
		t.Fatal(err)
	}

	// test

	exists, err := isExists(testPath)
	if err == nil {
		t.Error(err)
	}

	if exists {
		t.Errorf("Path %s should return `false` on error", testPath)
	}

	// tear down

	if err := os.Chmod(path.Dir(testPath), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(path.Dir(testPath)); err != nil {
		t.Error(err)
	}
}
