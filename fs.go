// This file provides an interface and mocks to os functions used in dotbro.

package main

import "os"

type DirMaker interface {
	MkdirAll(path string, perm os.FileMode) error
}

type FakeDirMaker struct {
	MkdirAllError error
}

func (f *FakeDirMaker) MkdirAll(path string, perm os.FileMode) error {
	return f.MkdirAllError
}

type Symlinker interface {
	Symlink(oldname, newname string) error
}

type FakeSymlinker struct {
	SymlinkError error
}

func (f *FakeSymlinker) Symlink(oldname, newname string) error {
	return f.SymlinkError
}

type Stater interface {
	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool
}

type MkdirSymlinker interface {
	DirMaker
	Symlinker
}

type FakeMkdirSymlinker struct {
	*FakeDirMaker
	*FakeSymlinker
}

type StatDirMaker interface {
	Stater
	DirMaker
}

// Actual implementation of interface using os package.

type OsDirMaker struct {
}

func (f *OsDirMaker) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

type OsSymlinker struct {
}

func (f *OsSymlinker) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

type OsMkdirSymlinker struct {
	*OsDirMaker
	*OsSymlinker
}
