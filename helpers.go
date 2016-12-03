package main

import (
	"os"
	"path/filepath"
)

// TODO: remove in favor of StatDirMaker
type DirCheckMaker interface {
	Stater
	MkdirAll(path string, perm os.FileMode) error
}

// CreatePath creates directories leading to file if they do not exist.
func CreatePath(dcm DirCheckMaker, file string) error {
	var err error
	var path = filepath.Dir(file)

	_, err = dcm.Stat(path)
	if err == nil {
		// path exists
		return nil
	}

	if !dcm.IsNotExist(err) {
		return err
	}

	return dcm.MkdirAll(path, 0700)
}

// IsExists reports whether path exists (path may be a file or a directory).
func IsExists(stater Stater, path string) (bool, error) {
	_, err := stater.Stat(path)
	if stater.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
