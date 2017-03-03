package main

import (
	//"io"
	"os"
	"time"
)

// FakeOS is a fake implementation of OS interface.
type FakeOS struct {
	OpenResult       File
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

func (f *FakeOS) Open(name string) (File, error) {
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

// FakeFile is kinda a os.File mock.
type FakeFile struct {
	CloseError    error
	StatResult    os.FileInfo
	StatError     error
	ReaddirResult []os.FileInfo
	ReaddirError  error
	ReadResult    int
	ReadError     error
}

func (f *FakeFile) Close() error {
	return f.CloseError
}

func (f *FakeFile) Stat() (os.FileInfo, error) {
	return f.StatResult, f.StatError
}

func (f *FakeFile) Readdir(n int) ([]os.FileInfo, error) {
	return f.ReaddirResult, f.ReaddirError
}

func (f *FakeFile) Read(p []byte) (n int, err error) {
	return f.ReadResult, f.ReadError
}

// FakeFileInfo is a os.FileInfo mock.
type FakeFileInfo struct {
	NameValue  string
	SizeValue  int64
	ModeValue  os.FileMode
	IsDirValue bool
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
	return f.IsDirValue
}

func (f *FakeFileInfo) Sys() interface{} {
	return nil
}
