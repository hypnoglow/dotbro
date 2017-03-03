package main

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

// TODO: the test is broken (cause of home)
func TestCleaner_CleanDeadSymlinks(t *testing.T) {
	cases := []struct {
		os            *FakeOS
		dirpath       string
		expectedError error
	}{
		{
			// cannot open dir
			os: &FakeOS{
				OpenError: errors.New("Cannot open dir"),
			},
			dirpath:       "/some/path",
			expectedError: errors.New("Cannot open dir"),
		},
		{
			// dir stat returned error
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatError: errors.New("Some error"),
				},
			},
			dirpath:       "/some/path",
			expectedError: errors.New("Some error"),
		},
		{
			// dir is not a dir
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: false,
					},
					ReaddirError: errors.New("Cannot read dir"),
				},
			},
			dirpath:       "/some/path",
			expectedError: errors.New("Specified dirPath /some/path is not a directory"),
		},
		{
			// dir readdir returned error
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: true,
					},
					ReaddirError: errors.New("Cannot read dir"),
				},
			},
			dirpath:       "/some/path",
			expectedError: errors.New("Cannot read dir"),
		},
		{
			// file is a dir
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: true,
					},
					ReaddirResult: []os.FileInfo{
						&FakeFileInfo{
							ModeValue: os.ModeDir,
						},
					},
				},
			},
			dirpath:       "/some/path",
			expectedError: nil,
		},
		{
			// file is a correct symlink
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: true,
					},
					ReaddirResult: []os.FileInfo{
						&FakeFileInfo{
							ModeValue: os.ModeSymlink,
						},
					},
				},
			},
			dirpath:       "/some/path",
			expectedError: nil,
		},
		{
			// file is a symlink, but error on stat
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: true,
					},
					ReaddirResult: []os.FileInfo{
						&FakeFileInfo{
							ModeValue: os.ModeSymlink,
						},
					},
				},
				StatError: os.ErrInvalid,
			},
			dirpath:       "/some/path",
			expectedError: os.ErrInvalid,
		},
		{
			// file is a symlink, but error on removal
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: true,
					},
					ReaddirResult: []os.FileInfo{
						&FakeFileInfo{
							ModeValue: os.ModeSymlink,
						},
					},
				},
				StatError:   os.ErrNotExist,
				RemoveError: os.ErrPermission,
			},
			dirpath:       "/some/path",
			expectedError: os.ErrPermission,
		},
		{
			// file is a symlink, successfully removed
			os: &FakeOS{
				OpenResult: &FakeFile{
					StatResult: &FakeFileInfo{
						IsDirValue: true,
					},
					ReaddirResult: []os.FileInfo{
						&FakeFileInfo{
							ModeValue: os.ModeSymlink,
						},
					},
				},
				StatError: os.ErrNotExist,
			},
			dirpath:       "/some/path",
			expectedError: nil,
		},
	}

	for _, c := range cases {
		cleaner := NewCleaner(&FakeOutputer{}, c.os)

		//// the hack
		//home, err := os.Open(os.ExpandEnv("$HOME"))
		//if err != nil {
		//	t.Fatal("Cannot open $HOME")
		//}
		//c.os.OpenResult = home

		err := cleaner.CleanDeadSymlinks(c.dirpath)

		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}

	}
}
