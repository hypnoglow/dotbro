package main

import (
	"os"
	"path"
	"path/filepath"
)

// processDest inspects destination path, and reports whether symlink and backup
// are needed.
func processDest(srcAbs string, destAbs string) (bool, bool, error) {
	var err error
	var needSymlink = false
	var needBackup = false

	fileInfo, err := os.Lstat(destAbs)
	if os.IsNotExist(err) {
		needSymlink = true
		return needSymlink, needBackup, nil
	}

	if err != nil {
		return needSymlink, needBackup, err
	}

	if fileInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
		needSymlink = true
		needBackup = true
		return needSymlink, needBackup, nil
	}

	target, err := os.Readlink(destAbs)
	if err != nil {
		return needSymlink, needBackup, err
	}

	if target == srcAbs {
		outVerbose("  ✓ %s is correct symlink", destAbs)
		return needSymlink, needBackup, nil
	}

	// dest is a wrong symlink
	// todo: if dry-run, just print
	err = os.Remove(destAbs)
	if err != nil {
		return needSymlink, needBackup, err
	}
	outInfo("  ✓ delete wrong symlink %s", destAbs)

	needSymlink = true
	return needSymlink, needBackup, nil
}

// backup copies existing destination file to backup dir.
func backup(dest string, destAbs string, backupDir string) error {
	// todo: if dry-run, just print

	dir := path.Dir(backupDir + "/" + dest)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	backupPath := backupDir + "/" + dest
	outVerbose("  → backup %s to %s", destAbs, backupPath)
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
	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	outVerbose("  → backup %s to %s", abs, backupPath)

	err = Copy(filename, backupPath)
	return err
}

// setSymlink symlinks scrAbs to destAbs
func setSymlink(srcAbs string, destAbs string) error {
	var err error

	// todo: if dry-run, just print

	dir := path.Dir(destAbs)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	err = os.Symlink(srcAbs, destAbs)
	if err != nil {
		return err
	}

	outInfo("  ✓ set symlink %s -> %s", srcAbs, destAbs)
	return nil
}
