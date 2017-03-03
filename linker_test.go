package main

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestLinker_Move(t *testing.T) {
	cases := []struct {
		os            *FakeOS
		oldpath       string
		newpath       string
		expectedError error
	}{
		{
			// Failure when IsExists fails
			os: &FakeOS{
				StatError: errors.New("Some error"),
			},
			expectedError: errors.New("Some error"),
		},
		{
			// Failure when dest file not exists
			os: &FakeOS{
				IsNotExistResult: true,
			},
			oldpath:       "/path/dest",
			expectedError: errors.New("File /path/dest not exists"),
		},
		{
			// Failure when MkdirAll fails
			os: &FakeOS{
				IsNotExistResult: false,
				MkdirAllError:    errors.New("MkdirAll error"),
			},
			expectedError: errors.New("MkdirAll error"),
		},
		{
			// Failure when Rename fails
			os: &FakeOS{
				IsNotExistResult: false,
				RenameError:      errors.New("Rename error"),
			},
			expectedError: errors.New("Rename error"),
		},
		{
			// Success
			os: &FakeOS{
				IsNotExistResult: false,
			},
			expectedError: nil,
		},
	}

	for _, c := range cases {
		linker := NewLinker(&FakeOutputer{}, c.os)
		err := linker.Move(c.oldpath, c.newpath)

		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}
	}
}

func TestLinker_SetSymlink(t *testing.T) {
	cases := []struct {
		os            *FakeOS
		srcAbs        string
		destAbs       string
		expectedError error
	}{
		{
			os: &FakeOS{
				MkdirAllError: nil,
				SymlinkError:  nil,
			},
			srcAbs:        "/src/path",
			destAbs:       "/dest/path",
			expectedError: nil,
		},
		{
			os: &FakeOS{
				MkdirAllError: errors.New("Permission denied"),
				SymlinkError:  nil,
			},
			srcAbs:        "/src/path",
			destAbs:       "/dest/path",
			expectedError: errors.New("Permission denied"),
		},
		{
			os: &FakeOS{
				MkdirAllError: nil,
				SymlinkError:  errors.New("File exists"),
			},
			srcAbs:        "/src/path",
			destAbs:       "/dest/path",
			expectedError: errors.New("File exists"),
		},
	}

	for _, c := range cases {
		linker := NewLinker(&FakeOutputer{}, c.os)

		err := linker.SetSymlink(c.srcAbs, c.destAbs)
		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}
	}

}

func TestLinker_NeedSymlink(t *testing.T) {
	cases := []struct {
		os             *FakeOS
		src            string
		dest           string
		expectedResult bool
		expectedError  error
	}{
		{
			os: &FakeOS{
				LstatError:       os.ErrNotExist,
				IsNotExistResult: true,
			},
			src:            "/src/path",
			dest:           "/some/non-existent/path",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			os: &FakeOS{
				LstatError: errors.New("Some error"),
			},
			src:            "/src/path",
			dest:           "/dest/path",
			expectedResult: false,
			expectedError:  errors.New("Some error"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: 0, // regular file
				},
			},
			src:            "/src/path",
			dest:           "/dest/path",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: os.ModeSymlink,
				},
				ReadlinkError: errors.New("Failed to read a link"),
			},
			src:            "/src/path",
			dest:           "/dest/path",
			expectedResult: false,
			expectedError:  errors.New("Failed to read a link"),
		},
		{
			// dest is already a correct symlink to src
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: os.ModeSymlink,
				},
				ReadlinkResult: "/src/path",
			},
			src:            "/src/path",
			dest:           "/dest/path",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			// dest is wrong symlink but failed to remove it
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: os.ModeSymlink,
				},
				ReadlinkResult: "/some/incorrect/path",
				RemoveError:    errors.New("Cannot remove"),
			},
			src:            "/src/path",
			dest:           "/dest/path",
			expectedResult: false,
			expectedError:  errors.New("Cannot remove"),
		},
		{
			// dest is wrong symlink and it removes successfully
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: os.ModeSymlink,
				},
				ReadlinkResult: "/some/incorrect/path",
				RemoveError:    nil,
			},
			src:            "/src/path",
			dest:           "/dest/path",
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, c := range cases {
		linker := NewLinker(&FakeOutputer{}, c.os)

		result, err := linker.NeedSymlink(c.src, c.dest)

		if result != c.expectedResult {
			t.Errorf("Expected %v but got %v\n", c.expectedResult, result)
		}

		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}

	}
}

func TestLinker_NeedBackup(t *testing.T) {
	cases := []struct {
		os             *FakeOS
		dest           string
		expectedResult bool
		expectedError  error
	}{
		{
			// dest path not exists
			os: &FakeOS{
				LstatError:       os.ErrNotExist,
				IsNotExistResult: true,
			},
			dest:           "/some/non-existent/path",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			// Lstat returned error
			os: &FakeOS{
				LstatError: errors.New("Some error"),
			},
			dest:           "/dest/path",
			expectedResult: false,
			expectedError:  errors.New("Some error"),
		},
		{
			// dest path is not a symlink
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: 0, // regular file
				},
			},
			dest:           "/dest/path",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			// dest path is a symlink
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					ModeValue: os.ModeSymlink,
				},
			},
			dest:           "/dest/path",
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, c := range cases {
		linker := NewLinker(&FakeOutputer{}, c.os)

		result, err := linker.NeedBackup(c.dest)

		if result != c.expectedResult {
			t.Errorf("Expected %v but got %v\n", c.expectedResult, result)
		}

		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}

	}
}
