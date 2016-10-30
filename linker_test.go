package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	returnCode := m.Run()
	os.RemoveAll("/tmp/dotbro") // Cleanup
	os.Exit(returnCode)
}

func TestNeedSymlink(t *testing.T) {
	// Test dest does not exist
	src := "/tmp/dotbro/linker/original.txt"
	dest := "/tmp/dotbro/linker/original.txt"

	actual, err := NeedSymlink(src, dest)
	assert.Equal(t, actual, true)
	assert.Equal(t, err, nil)

	// Test destination is not a symlink
	if err = os.MkdirAll(path.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(src, nil, 0333); err != nil {
		t.Fatal(err)
	}
	actual, err = NeedSymlink(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, actual, true)

	dest = "/tmp/dotbro/linker/original"
	if err = os.Symlink(src, dest); err != nil {
		t.Fatal(err)
	}

	// Test destination is a symlink
	actual, err = NeedSymlink(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, actual, false)

}
