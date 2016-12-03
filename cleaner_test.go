package main

import "testing"

func TestCleanDeadSymlinks(t *testing.T) {
	var dirpath string
	var err error

	dirpath = "/some/nonexistent/path"
	err = cleanDeadSymlinks(dirpath)
	if err == nil {
		t.Errorf("[FAIL] cleanDeadSymlinks(%q) -> should error\n", dirpath)
	} else {
		t.Logf("[OK] cleanDeadSymlinks(%q) -> errored with %q\n", dirpath, err)
	}

	dirpath = "/home/hypnoglow/.vimrc"
	err = cleanDeadSymlinks(dirpath)
	if err == nil {
		t.Errorf("[FAIL] cleanDeadSymlinks(%q) -> should error\n", dirpath)
	} else {
		t.Logf("[OK] cleanDeadSymlinks(%q) -> errored with %q\n", dirpath, err)
	}

	dirpath = "/tmp"
	err = cleanDeadSymlinks(dirpath)
	if err != nil {
		t.Errorf("[FAIL] cleanDeadSymlinks(%q) -> should not error, but errored: %q\n", dirpath, err)
	} else {
		t.Logf("[OK] cleanDeadSymlinks(%q) -> should not error\n", dirpath)
	}
}
