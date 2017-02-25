// This file provides an interface and mocks to os functions used in dotbro.

package main

import (
	"os"
	"time"
)

// Interfaces

type OS interface {
	Open(name string) (*os.File, error)
	Create(name string) (*os.File, error)

	MkdirAll(path string, perm os.FileMode) error

	Symlink(oldname, newname string) error

	Stat(name string) (os.FileInfo, error)
	Lstat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool

	Rename(oldpath, newpath string) error
}

// Fake implementation of interface

type FakeOS struct {
	OpenResult       *os.File
	OpenError        error
	CreateResult     *os.File
	CreateError      error
	MkdirAllError    error
	SymlinkError     error
	StatFileInfo     os.FileInfo
	StatError        error
	LstatFileInfo    os.FileInfo
	LstatError       error
	IsNotExistResult bool
	RenameError      error
}

func (f *FakeOS) Open(name string) (*os.File, error) {
	return f.OpenResult, f.OpenError
}

func (f *FakeOS) Create(name string) (*os.File, error) {
	return f.CreateResult, f.CreateError
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

func (f *FakeOS) Lstat(name string) (os.FileInfo, error) {
	return f.LstatFileInfo, f.LstatError
}

func (f *FakeOS) IsNotExist(err error) bool {
	return f.IsNotExistResult
}

func (f *FakeOS) Rename(oldname, newname string) error {
	return f.RenameError
}

// Actual implementation of interface using os package.

type OSFS struct{}

func (f *OSFS) Open(name string) (*os.File, error) {
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

// FileInfo mock

type FakeFileInfo struct {
	NameValue string
	SizeValue int64
	ModeValue os.FileMode
}

func (f *FakeFileInfo) Name() string {
	return f.NameValue
}

func (f *FakeFileInfo) Size() int64 {
	return f.SizeValue
}

func (f *FakeFileInfo) Mode() os.FileMode {
	return f.ModeValue
}

func (f *FakeFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (f *FakeFileInfo) IsDir() bool {
	return false
}

func (f *FakeFileInfo) Sys() interface{} {
	return nil
}
