package main

import (
	"fmt"
	"os"
	"path"

	. "github.com/logrusorgru/aurora"
)

type Linker struct {
	log LevelLogger
	os  OS
}

func NewLinker(outputer LevelLogger, os OS) Linker {
	return Linker{
		log: outputer,
		os:  os,
	}
}

// Move moves oldpath to newpath, creating target directories if need.
func (l *Linker) Move(oldpath, newpath string) error {
	// check if oldpath file exists
	_, err := l.os.Stat(oldpath)
	if l.os.IsNotExist(err) {
		return fmt.Errorf("File %s not exists", oldpath)
	}
	if err != nil {
		return err
	}

	err = l.os.MkdirAll(path.Dir(newpath), 0700)
	if err != nil {
		return err
	}

	l.log.Debug("  %s backup %s to %s", Green("→"), Brown(oldpath), Brown(newpath))
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

	l.log.Info("  %s set symlink %s -> %s", Green("+"), Brown(srcAbs), Brown(destAbs))
	return nil
}

// NeedSymlink reports whether source file needs to be symlinked to destination path.
func (l *Linker) NeedSymlink(src, dest string) (bool, error) {
	fi, err := l.os.Lstat(dest)
	if l.os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
		return true, nil
	}

	target, err := l.os.Readlink(dest)
	if err != nil {
		return false, err
	}

	if target == src {
		l.log.Debug("  %s %s is correct symlink", Green("✓"), Brown(dest))
		return false, nil
	}

	// here dest is a wrong symlink

	if err = l.os.Remove(dest); err != nil {
		return false, err
	}
	l.log.Info("  %s delete wrong symlink %s", Green("✓"), Brown(dest))

	return true, nil
}

// NeedBackup reports whether destination path needs to be backed up.
func (l *Linker) NeedBackup(dest string) (bool, error) {
	fi, err := l.os.Lstat(dest)
	if l.os.IsNotExist(err) {
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
