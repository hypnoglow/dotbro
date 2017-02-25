// This file provides an interface and mocks to os functions used in dotbro.

package main

import "os"

// Interfaces

type OS interface {
	MkdirAll(path string, perm os.FileMode) error

	Symlink(oldname, newname string) error

	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool

	Rename(oldpath, newpath string) error
}

// Fake implementation of interface

type FakeOS struct {
	MkdirAllError    error
	SymlinkError     error
	StatFileInfo     os.FileInfo
	StatError        error
	IsNotExistResult bool
	RenameError      error
}

func (f *FakeOS) MkdirAll(path string, perm os.FileMode) error {
	return f.MkdirAllError
}

func (f *FakeOS) Symlink(oldname, newname string) error {
	return f.SymlinkError
}

func (f *FakeOS) Stat(name string) (os.FileInfo, error) {
	return f.StatFileInfo, f.StatError
}

func (f *FakeOS) IsNotExist(err error) bool {
	return f.IsNotExistResult
}

func (f *FakeOS) Rename(oldname, newname string) error {
	return f.RenameError
}

// Actual implementation of interface using os package.

type OSFS struct{}

func (f *OSFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *OSFS) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (f *OSFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *OSFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (f *OSFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
