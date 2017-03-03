// This file provides an interface to os functions used in dotbro.

package main

import "os"

// Interfaces

type OS interface {
	Open(name string) (File, error)
	Create(name string) (*os.File, error)

	MkdirAll(path string, perm os.FileMode) error

	Symlink(oldname, newname string) error
	Readlink(name string) (string, error)

	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool

	Rename(oldpath, newpath string) error
	Remove(name string) error
}

type File interface {
	Close() error
	Stat() (os.FileInfo, error)
	Readdir(n int) ([]os.FileInfo, error)
	Read(p []byte) (n int, err error)
}

// Actual implementation of interface using os package.

type OSFS struct{}

func (f *OSFS) Open(name string) (File, error) {
	return os.Open(name)
}

func (f *OSFS) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (f *OSFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *OSFS) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (f *OSFS) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (f *OSFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (f *OSFS) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func (f *OSFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (f *OSFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (f *OSFS) Remove(name string) error {
	return os.Remove(name)
}
