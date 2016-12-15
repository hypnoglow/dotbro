package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestIsExists(t *testing.T) {
	cases := []struct {
		stater         *FakeStater
		name           string
		expectedResult bool
		expectedError  error
	}{
		{
			stater: &FakeStater{
				StatFileInfo:     nil, // does not matter
				StatError:        errors.New("Permission denied"),
				IsNotExistResult: false,
			},
			name:           "/path/that/errors/on/stat",
			expectedResult: false,
			expectedError:  errors.New("Permission denied"),
		},
		{
			stater: &FakeStater{
				StatFileInfo:     nil, // does not matter
				StatError:        errors.New("Not exists"),
				IsNotExistResult: true,
			},
			name:           "/path/that/exists",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			stater: &FakeStater{
				StatFileInfo:     nil, // does not matter
				StatError:        nil,
				IsNotExistResult: false,
			},
			name:           "/path/that/not/exists",
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, testcase := range cases {
		exists, err := IsExists(testcase.stater, testcase.name)
		if exists != testcase.expectedResult {
			t.Errorf("Expected %v but got %v\n", testcase.expectedResult, exists)
		}

		if !reflect.DeepEqual(err, testcase.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", testcase.expectedError, err)
		}
	}
}

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
