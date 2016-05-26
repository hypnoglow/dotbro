package main

import (
	"os"
	"path/filepath"
)

// createPath creates directories leading to file if they do not exist.
func createPath(file string) error {
	var err error
	var path = filepath.Dir(file)

	_, err = os.Stat(path)
	if err == nil {
		// path exists
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return nil
}

// isExists reports whether path exists (path may be a file or a directory).
func isExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
