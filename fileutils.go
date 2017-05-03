package main

import (
	"fmt"
	"io"
	"path"
)

// Copy copies a file from src to dst.
// src and dest can be either absolute or relative paths.
func Copy(osfs OS, src, dst string) error {
	sfi, err := osfs.Lstat(src)
	if err != nil {
		return err
	}

	if !sfi.Mode().IsRegular() {
		return fmt.Errorf("Non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	dfi, err := osfs.Stat(dst)
	if err != nil {
		if !osfs.IsNotExist(err) {
			return err
		}
		// file not exists - do not do anything
	} else {
		// file exists - check it
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("Non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
	}

	err = copyFileContents(osfs, src, dst)
	return err
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(osfs OS, src, dst string) (err error) {
	in, err := osfs.Open(src)
	if err != nil {
		return
	}

	defer func() {
		inCloseErr := in.Close()
		if err == nil {
			err = inCloseErr
		}
	}()

	if err = osfs.MkdirAll(path.Dir(dst), 0700); err != nil {
		return err
	}

	out, err := osfs.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		outCloseErr := out.Close()
		if err == nil {
			err = outCloseErr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	err = out.Sync()
	return err
}
