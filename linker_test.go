package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeLogWriterForLinkerOutputer struct{}

func (f *FakeLogWriterForLinkerOutputer) Write(format string, v ...interface{}) {
	return
}

//
// func TestMain(m *testing.M) {
// 	// returnCode := m.Run()
// 	os.RemoveAll("/tmp/dotbro") // Cleanup
// 	// os.Exit(returnCode)
// }

func TestNeedSymlink(t *testing.T) {
	// TODO: test fails if outputer is not defined.
	outputer = NewOutputer(OutputerModeQuiet, os.Stdout, &FakeLogWriterForLinkerOutputer{})

	os.RemoveAll("/tmp/dotbro") // Cleanup

	// Test dest does not exist
	src := "/tmp/dotbro/linker/TestNeedSymlink.txt"
	dest := "/tmp/dotbro/linker/TestNeedSymlink.txt"
	wrongDest := "/tmp/dotbro/linker/wrongTestNeedSymlink"

	actual, err := needSymlink(src, dest)
	assert.True(t, actual)
	assert.Equal(t, err, nil)

	// Test destination is not a symlink
	if err = os.MkdirAll(path.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(src, nil, 0333); err != nil {
		t.Fatal(err)
	}
	actual, err = needSymlink(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, actual)

	dest = "/tmp/dotbro/linker/TestNeedSymlink"
	if err = os.Symlink(src, dest); err != nil {
		t.Fatal(err)
	}

	// Test destination is a symlink
	actual, err = needSymlink(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, actual)

	// Test symlink goes to the wrong destination
	if err = os.Remove(dest); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(wrongDest, dest); err != nil {
		t.Fatal(err)
	}
	actual, err = needSymlink(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, actual)

}

func TestNeedBackup(t *testing.T) {

	os.RemoveAll("/tmp/dotbro") // Cleanup

	// Test dest does not exist
	dest := "/tmp/dotbro/linker/TestNeedBackup.txt"

	actual, err := needBackup(dest)
	assert.False(t, actual)
	assert.Empty(t, err)

	// Test destination is not a symlink
	src := "/tmp/dotbro/linker/TestNeedBackup.txt"
	if err = os.MkdirAll(path.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(src, nil, 0333); err != nil {
		t.Fatal(err)
	}
	actual, err = needBackup(dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, actual)

	dest = "/tmp/dotbro/linker/TestNeedBackup"
	if err = os.Symlink(src, dest); err != nil {
		t.Fatal(err)
	}

	// Test destination is a symlink
	actual, err = needBackup(dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, actual)
}

func TestMove(t *testing.T) {
	cases := []struct {
		os            *FakeOS
		oldpath       string
		newpath       string
		expectedError error
	}{
		{
			// Failure when IsExists fails
			os: &FakeOS{
				StatError: errors.New("Some error"),
			},
			expectedError: errors.New("Some error"),
		},
		{
			// Failure when dest file not exists
			os: &FakeOS{
				IsNotExistResult: true,
			},
			oldpath:       "/path/dest",
			expectedError: errors.New("File /path/dest not exists"),
		},
		{
			// Failure when MkdirAll fails
			os: &FakeOS{
				IsNotExistResult: false,
				MkdirAllError:    errors.New("MkdirAll error"),
			},
			expectedError: errors.New("MkdirAll error"),
		},
		{
			// Failure when Rename fails
			os: &FakeOS{
				IsNotExistResult: false,
				RenameError:      errors.New("Rename error"),
			},
			expectedError: errors.New("Rename error"),
		},
		{
			// Success
			os: &FakeOS{
				IsNotExistResult: false,
			},
			expectedError: nil,
		},
	}

	for _, c := range cases {
		linker := NewLinker(&FakeOutputer{}, c.os)
		err := linker.Move(c.oldpath, c.newpath)

		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}
	}
}

func TestNewLinker(t *testing.T) {
	cases := []struct {
		os            *FakeOS
		srcAbs        string
		destAbs       string
		expectedError error
	}{
		{
			os: &FakeOS{
				MkdirAllError: nil,
				SymlinkError:  nil,
			},
			srcAbs:        "/src/path",
			destAbs:       "/dest/path",
			expectedError: nil,
		},
		{
			os: &FakeOS{
				MkdirAllError: errors.New("Permission denied"),
				SymlinkError:  nil,
			},
			srcAbs:        "/src/path",
			destAbs:       "/dest/path",
			expectedError: errors.New("Permission denied"),
		},
		{
			os: &FakeOS{
				MkdirAllError: nil,
				SymlinkError:  errors.New("File exists"),
			},
			srcAbs:        "/src/path",
			destAbs:       "/dest/path",
			expectedError: errors.New("File exists"),
		},
	}

	for _, c := range cases {
		linker := NewLinker(&FakeOutputer{}, c.os)

		err := linker.SetSymlink(c.srcAbs, c.destAbs)
		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}
	}

}
