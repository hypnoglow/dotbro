package main

import (
	"fmt"
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
	src := "/tmp/dotbro/linker/TestNeedSymlink.txt"
	dest := "/tmp/dotbro/linker/TestNeedSymlink.txt"
	wrongDest := "/tmp/dotbro/linker/wrongTestNeedSymlink"

	actual, err := needSymlink(src, dest)
	assert.True(t, true)
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

func TestBackup(t *testing.T) {
	dest := "new"
	destAbs := "/tmp/dotbro/linker/TestBackup/new"
	backupDir := "/tmp/dotbro/linker/TestBackup/backup"

	err := backup(dest, destAbs, backupDir)
	assert.Error(t, err)

	err = os.MkdirAll(destAbs, 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = backup(dest, destAbs, backupDir)
	assert.Empty(t, err)
}

func TestBackupCopy(t *testing.T) {
	filename := "/tmp/dotbro/linker/TestBackupCopy/file"
	backupDir := "/tmp/dotbro/linker/TestBackupCopy/backup"
	if err := os.MkdirAll(path.Dir(filename), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filename, []byte("TestData"), 0755); err != nil {
		t.Fatal(err)
	}

	err := backupCopy(filename, backupDir)
	assert.Empty(t, err)
	filenameContent, _ := ioutil.ReadFile(filename)
	backupContent, _ := ioutil.ReadFile(backupDir + "/file")
	assert.Equal(t, filenameContent, backupContent)
}

func TestSetSymlink(t *testing.T) {
	srcAbs := "/tmp/dotbro/linker/TestSetSymlink/file"
	destAbs := "/tmp/dotbro/linker/TestSetSymlink/filesymlink"
	err := setSymlink(srcAbs, destAbs)
	assert.Nil(t, err)

	// Calling again should return an error since the link already exists
	err = setSymlink(srcAbs, destAbs)
	assert.EqualError(t, err, fmt.Sprintf("symlink %s %s: file exists", srcAbs, destAbs))
}
