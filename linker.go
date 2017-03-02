package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type Linker struct {
	outputer IOutputer
	os       OS
}

func NewLinker(outputer IOutputer, os OS) Linker {
	return Linker{
		outputer: outputer,
		os:       os,
	}
}

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

// Move moves oldpath to newpath, creating target directories if need.
func (l *Linker) Move(oldpath, newpath string) error {
	// todo: if dry-run, just print

	// check if destination file exists
	exists, err := IsExists(l.os, oldpath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("File %s not exists", oldpath)
	}

	err = l.os.MkdirAll(path.Dir(newpath), 0700)
	if err != nil {
		return err
	}

	l.outputer.OutVerbose("  → backup %s to %s", oldpath, newpath)
	err = l.os.Rename(oldpath, newpath)
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

	// TODO
	err = Copy(osfs, filename, backupPath)
	return err
}

// SetSymlink symlinks scrAbs to destAbs.
func (l *Linker) SetSymlink(srcAbs string, destAbs string) error {

	dir := path.Dir(destAbs)
	if err := l.os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if err := l.os.Symlink(srcAbs, destAbs); err != nil {
		return err
	}

	l.outputer.OutInfo("  ✓ set symlink %s -> %s", srcAbs, destAbs)
	return nil
}
