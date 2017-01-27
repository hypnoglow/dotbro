// This file provides an interface and mocks to os functions used in dotbro.

package main

import "os"

// Intefaces

type OS interface {
	MkdirAll(path string, perm os.FileMode) error

	Symlink(oldname, newname string) error

	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool

	Rename(oldpath, newpath string) error
}

type DirMaker interface {
	MkdirAll(path string, perm os.FileMode) error
}

type Symlinker interface {
	Symlink(oldname, newname string) error
}

type Stater interface {
	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool
}

type MkdirSymlinker interface {
	DirMaker
	Symlinker
}

type StatDirMaker interface {
	Stater
	DirMaker
}

type Renamer interface {
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

type FakeDirMaker struct {
	MkdirAllError error
}

func (f *FakeDirMaker) MkdirAll(path string, perm os.FileMode) error {
	return f.MkdirAllError
}

type FakeSymlinker struct {
	SymlinkError error
}

func (f *FakeSymlinker) Symlink(oldname, newname string) error {
	return f.SymlinkError
}

type FakeStater struct {
	StatFileInfo     os.FileInfo
	StatError        error
	IsNotExistResult bool
}

func (f *FakeStater) Stat(name string) (os.FileInfo, error) {
	return f.StatFileInfo, f.StatError
}

func (f *FakeStater) IsNotExist(err error) bool {
	return f.IsNotExistResult
}

type FakeMkdirSymlinker struct {
	*FakeDirMaker
	*FakeSymlinker
}

type FakeStatDirMaker struct {
	*FakeStater
	*FakeDirMaker
}

type FakeRenamer struct {
	RenameError error
}

func (f *FakeRenamer) Rename(oldpath, newpath string) error {
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

func (s *OSFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (s *OSFS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (f *OSFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

type OsDirMaker struct {
}

func (f *OsDirMaker) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

type OsSymlinker struct{}

func (f *OsSymlinker) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

type OsStater struct{}

func (s *OsStater) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (s *OsStater) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

type OsMkdirSymlinker struct {
	*OsDirMaker
	*OsSymlinker
}

type OsStatDirMaker struct {
	*OsStater
	*OsDirMaker
}

type OsRenamer struct{}

func (f *OsRenamer) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
