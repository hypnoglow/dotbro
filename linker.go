package main

import (
	"os"
	"path"
	"path/filepath"
)

// needSymlink reports whether source file needs to be symlinked to destination path.
func needSymlink(src, dest string) (bool, error) {
	fi, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
		return true, nil
	}

	target, err := os.Readlink(dest)
	if err != nil {
		return false, err
	}

	if target == src {
		outputer.OutVerbose("  ✓ %s is correct symlink", dest)
		return false, nil
	}

	// here dest is a wrong symlink

	// todo: if dry-run, just print
	err = os.Remove(dest)
	if err != nil {
		return false, err
	}
	outputer.OutInfo("  ✓ delete wrong symlink %s", dest)

	return true, nil
}

// needBackup reports whether destination path needs to be backed up.
func needBackup(dest string) (bool, error) {
	fi, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
		return true, nil
	}

	return false, nil
}

// backup copies existing destination file to backup dir.
func backup(dest string, destAbs string, backupDir string) error {
	// todo: if dry-run, just print

	dir := path.Dir(backupDir + "/" + dest)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	backupPath := backupDir + "/" + dest
	outputer.OutVerbose("  → backup %s to %s", destAbs, backupPath)
	err = os.Rename(destAbs, backupPath)
	return err
}

func backupCopy(filename, backupDir string) error {
	rel := path.Base(filename)
	abs, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	backupPath := backupDir + "/" + rel

	// Create subdirectories, if need
	dir := path.Dir(backupPath)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	outputer.OutVerbose("  → backup %s to %s", abs, backupPath)

	err = Copy(filename, backupPath)
	return err
}

// setSymlink symlinks scrAbs to destAbs
func setSymlink(srcAbs string, destAbs string) error {
	var err error

	// todo: if dry-run, just print

	dir := path.Dir(destAbs)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	err = os.Symlink(srcAbs, destAbs)
	if err != nil {
		return err
	}

	outputer.OutInfo("  ✓ set symlink %s -> %s", srcAbs, destAbs)
	return nil
}
