package main

import (
	"os"
	"time"
)

// FakeOS is a fake implementation of OS interface.
type FakeOS struct {
	OpenResult       *os.File
	OpenError        error
	CreateResult     *os.File
	CreateError      error
	MkdirAllError    error
	SymlinkError     error
	ReadlinkResult   string
	ReadlinkError    error
	StatFileInfo     os.FileInfo
	StatError        error
	LstatFileInfo    os.FileInfo
	LstatError       error
	IsNotExistResult bool
	RenameError      error
	RemoveError      error
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

func (f *FakeOS) Readlink(name string) (string, error) {
	return f.ReadlinkResult, f.ReadlinkError
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

func (f *FakeOS) Remove(name string) error {
	return f.RemoveError
}

// FakeFileInfo is a os.FileInfo mock.
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
