package main

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

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
			// cannot stat file to clean
			os: &FakeOS{
				StatError: errors.New("Permission denied"),
			},
			dirpath:       "/some/path",
			expectedError: errors.New("Permission denied"),
		},
		{
			// correct file
			os: &FakeOS{
				StatError: nil,
			},
			dirpath:       "/some/path",
			expectedError: nil,
		},
		{
			// cannot remove file
			os: &FakeOS{
				StatError:   os.ErrNotExist,
				RemoveError: errors.New("Cannot remove file"),
			},
			dirpath:       "/some/path",
			expectedError: errors.New("Cannot remove file"),
		},
		{
			// successful cleaning
			os: &FakeOS{
				StatError:   os.ErrNotExist,
				RemoveError: nil,
			},
			dirpath:       "/some/path",
			expectedError: nil,
		},
	}

	for _, c := range cases {
		cleaner := NewCleaner(&FakeOutputer{}, c.os)

		// the hack
		home, err := os.Open(os.ExpandEnv("$HOME"))
		if err != nil {
			t.Fatal("Cannot open $HOME")
		}
		c.os.OpenResult = home

		err = cleaner.CleanDeadSymlinks(c.dirpath)

		if !reflect.DeepEqual(err, c.expectedError) {
			t.Errorf("Expected err to be %v but it was %v\n", c.expectedError, err)
		}

	}
}
