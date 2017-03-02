package main

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestIsExists(t *testing.T) {
	cases := []struct {
		os             *FakeOS
		name           string
		expectedResult bool
		expectedError  error
	}{
		{
			os: &FakeOS{
				StatFileInfo:     nil, // does not matter
				StatError:        errors.New("Permission denied"),
				IsNotExistResult: false,
			},
			name:           "/path/that/errors/on/stat",
			expectedResult: false,
			expectedError:  errors.New("Permission denied"),
		},
		{
			os: &FakeOS{
				StatFileInfo:     nil, // does not matter
				StatError:        errors.New("Not exists"),
				IsNotExistResult: true,
			},
			name:           "/path/that/exists",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			os: &FakeOS{
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
		exists, err := IsExists(testcase.os, testcase.name)
		if exists != testcase.expectedResult {
			t.Errorf("Expected %v but got %v\n", testcase.expectedResult, exists)
		}

		if !reflect.DeepEqual(err, testcase.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", testcase.expectedError, err)
		}
	}
}

func TestCopyReal(t *testing.T) {
	cases := []struct {
		os            *FakeOS
		src           string
		dest          string
		expectedError error
	}{
		{
			os: &FakeOS{
				LstatError:       errors.New("Permission denied"),
				IsNotExistResult: false,
			},
			src:           "/path/to/source",
			dest:          "/path/to/dest",
			expectedError: errors.New("Permission denied"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "source",
					ModeValue: os.ModeDir,
				},
			},
			src:           "/path/to/source",
			dest:          "/path/to/dest",
			expectedError: errors.New("Non-regular source file source (\"d---------\")"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "source",
					ModeValue: 0,
				},
				StatError: errors.New("Permisson denied 123"),
			},
			src:           "/path/to/source",
			dest:          "/path/to/dest",
			expectedError: errors.New("Permisson denied 123"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "source",
					ModeValue: 0,
				},
				StatFileInfo: &FakeFileInfo{
					NameValue: "dest",
					ModeValue: os.ModeDir,
				},
			},
			src:           "/path/to/source",
			dest:          "/path/to/dest",
			expectedError: errors.New("Non-regular destination file dest (\"d---------\")"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "sourcefile",
					ModeValue: 0,
				},
				StatFileInfo: &FakeFileInfo{
					NameValue: "destfile",
					ModeValue: 0,
				},
				OpenError: errors.New("Cannot open file sourcefile"),
			},
			src:           "/path/to/sourcefile",
			dest:          "/path/to/destfile",
			expectedError: errors.New("Cannot open file sourcefile"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "source",
					ModeValue: 0,
				},
				StatFileInfo: &FakeFileInfo{
					NameValue: "dest",
					ModeValue: 0,
				},
				MkdirAllError: errors.New("Cannot create dir"),
			},
			src:           "/path/to/source",
			dest:          "/path/to/dest",
			expectedError: errors.New("Cannot create dir"),
		},
		{
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "sourcefile",
					ModeValue: 0,
				},
				StatFileInfo: &FakeFileInfo{
					NameValue: "destfile",
					ModeValue: 0,
				},
				CreateError: errors.New("Cannot create file destfile"),
			},
			src:           "/path/to/sourcefile",
			dest:          "/path/to/destfile",
			expectedError: errors.New("Cannot create file destfile"),
		},
	}

	for _, testcase := range cases {
		err := Copy(testcase.os, testcase.src, testcase.dest)

		if !reflect.DeepEqual(err, testcase.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", testcase.expectedError, err)
		}
	}
}
