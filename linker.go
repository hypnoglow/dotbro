package main

import (
	"fmt"
	"os"
	"path"

	. "github.com/logrusorgru/aurora"
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

// Move moves oldpath to newpath, creating target directories if need.
func (l *Linker) Move(oldpath, newpath string) error {
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

	l.outputer.OutVerbose("  %s backup %s to %s", Green("→"), oldpath, newpath)
	err = l.os.Rename(oldpath, newpath)
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

	l.outputer.OutInfo("  %s set symlink %s -> %s", Green("+"), Brown(srcAbs), Brown(destAbs))
	return nil
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
		outputer.OutVerbose("  %s %s is correct symlink", Green("✓"), Brown(dest))
		return false, nil
	}

	// here dest is a wrong symlink

	// todo: if dry-run, just print
	err = os.Remove(dest)
	if err != nil {
		return false, err
	}
	outputer.OutInfo("  %s delete wrong symlink %s", Green("✓"), Brown(dest))

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
