package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
)

type Linker struct {
	os     OS
	logger *slog.Logger
}

func NewLinker(os OS, logger *slog.Logger) Linker {
	return Linker{
		os:     os,
		logger: logger,
	}
}

// Move moves oldpath to newpath, creating target directories if need.
func (l *Linker) Move(ctx context.Context, oldpath, newpath string) error {
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

	l.logger.DebugContext(ctx, "backup",
		slog.String("status", "→"),
		slog.String("src", oldpath),
		slog.String("dst", newpath))
	err = l.os.Rename(oldpath, newpath)
	return err
}

// SetSymlink symlinks scrAbs to destAbs.
func (l *Linker) SetSymlink(srcAbs string, destAbs string) error {

	dir := path.Dir(destAbs)
	if err := l.os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return l.os.Symlink(srcAbs, destAbs)
}

// NeedSymlink reports whether source file needs to be symlinked to destination path.
func (l *Linker) NeedSymlink(ctx context.Context, src, dest string) (bool, error) {
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
		l.logger.DebugContext(ctx, "correct symlink",
			slog.String("status", "✓"),
			slog.String("path", dest))
		return false, nil
	}

	// here dest is a wrong symlink

	if err = l.os.Remove(dest); err != nil {
		return false, err
	}
	l.logger.InfoContext(ctx, "delete wrong symlink",
		slog.String("status", "✓"),
		slog.String("path", dest))

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
