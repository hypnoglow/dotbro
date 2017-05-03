package main

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestCopy(t *testing.T) {
	// dirty hacks
	testLonelyFile, err := os.Create("/tmp/test_in_file")
	if err != nil {
		t.Fatal(err)
	}
	testInFile, err := os.Create("/tmp/test_in_file")
	if err != nil {
		t.Fatal(err)
	}
	testOutFile, err := os.Create("/tmp/test_out_file")
	if err != nil {
		t.Fatal(err)
	}

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
				OpenResult:    os.Stdin, // just in case
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
				OpenResult:  os.Stdin, // just in case
				CreateError: errors.New("Cannot create file destfile"),
			},
			src:           "/path/to/sourcefile",
			dest:          "/path/to/destfile",
			expectedError: errors.New("Cannot create file destfile"),
		},
		{
			// no out file
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "sourcefile",
					ModeValue: 0,
				},
				StatFileInfo: &FakeFileInfo{
					NameValue: "destfile",
					ModeValue: 0,
				},
				OpenResult: testLonelyFile,
			},
			src:           "/path/to/sourcefile",
			dest:          "/path/to/destfile",
			expectedError: errors.New("invalid argument"),
		},
		{
			// all is ok
			os: &FakeOS{
				LstatFileInfo: &FakeFileInfo{
					NameValue: "sourcefile",
					ModeValue: 0,
				},
				StatFileInfo: &FakeFileInfo{
					NameValue: "destfile",
					ModeValue: 0,
				},
				OpenResult:   testInFile,
				CreateResult: testOutFile,
			},
			src:           "/path/to/sourcefile",
			dest:          "/path/to/destfile",
			expectedError: nil,
		},
	}

	for _, testcase := range cases {
		err := Copy(testcase.os, testcase.src, testcase.dest)

		if !reflect.DeepEqual(err, testcase.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", testcase.expectedError, err)
		}
	}
}
