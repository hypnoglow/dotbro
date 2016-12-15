// This file provides an interface and mocks to os functions used in dotbro.

package main

import "os"

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

type FakeMkdirSymlinker struct {
	*FakeDirMaker
	*FakeSymlinker
}

type StatDirMaker interface {
	Stater
	DirMaker
}

// Fake implementation of interface

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

type MkdirSymlinker interface {
	DirMaker
	Symlinker
}

// Actual implementation of interface using os package.

type OsDirMaker struct {
}

func (f *OsDirMaker) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

type OsSymlinker struct{}

func (f *OsSymlinker) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

type OsMkdirSymlinker struct {
	*OsDirMaker
	*OsSymlinker
}

type OsStater struct{}

func (s *OsStater) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (s *OsStater) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
