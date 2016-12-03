package main

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

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

type FakeDirCheckMaker struct {
	*FakeStater
	MkdirAllError error
}

func (f *FakeDirCheckMaker) MkdirAll(path string, perm os.FileMode) error {
	return f.MkdirAllError
}

func TestIsExists(t *testing.T) {
	cases := []struct {
		stater         *FakeStater
		name           string
		expectedResult bool
		expectedError  error
	}{
		{
			stater: &FakeStater{
				StatFileInfo:     nil, // does not matter
				StatError:        errors.New("Permission denied"),
				IsNotExistResult: false,
			},
			name:           "/path/that/errors/on/stat",
			expectedResult: false,
			expectedError:  errors.New("Permission denied"),
		},
		{
			stater: &FakeStater{
				StatFileInfo:     nil, // does not matter
				StatError:        errors.New("Not exists"),
				IsNotExistResult: true,
			},
			name:           "/path/that/exists",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			stater: &FakeStater{
				StatFileInfo:     nil, // does not matter
				StatError:        nil,
				IsNotExistResult: false,
			},
			name:           "/path/that/not/exists",
			expectedResult: true,
			expectedError:  nil,
		},
	}

	for _, testcase := range cases {
		exists, err := IsExists(testcase.stater, testcase.name)
		if exists != testcase.expectedResult {
			t.Errorf("Expected %v but got %v\n", testcase.expectedResult, exists)
		}

		if !reflect.DeepEqual(err, testcase.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", testcase.expectedError, err)
		}
	}
}

func TestCreatePath(t *testing.T) {
	cases := []struct {
		dirCheckMaker *FakeDirCheckMaker
		file          string
		expectedError error
	}{
		{
			dirCheckMaker: &FakeDirCheckMaker{
				FakeStater: &FakeStater{
					StatError: nil,
				},
			},
			file:          "/path/that/already/exists",
			expectedError: nil,
		},
		{
			dirCheckMaker: &FakeDirCheckMaker{
				FakeStater: &FakeStater{
					StatError:        errors.New("Not exists"),
					IsNotExistResult: true,
				},
			},
			file:          "/path/that/will/be/created",
			expectedError: nil,
		},
		{
			dirCheckMaker: &FakeDirCheckMaker{
				FakeStater: &FakeStater{
					StatError:        errors.New("Permission denied"),
					IsNotExistResult: false,
				},
				MkdirAllError: nil,
			},
			file:          "/path/that/cannot/be/scanned",
			expectedError: errors.New("Permission denied"),
		},
		{
			dirCheckMaker: &FakeDirCheckMaker{
				FakeStater: &FakeStater{
					StatError:        errors.New("Not exists"),
					IsNotExistResult: true,
				},
				MkdirAllError: errors.New("Cannot create dir: Permission denied"),
			},
			file:          "/path/that/cannot/be/created",
			expectedError: errors.New("Cannot create dir: Permission denied"),
		},
	}

	for _, testcase := range cases {
		err := CreatePath(testcase.dirCheckMaker, testcase.file)

		if !reflect.DeepEqual(err, testcase.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", testcase.expectedError, err)
		}
	}
}
