package main

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestNewProfile(t *testing.T) {
	NewProfile()
}

func TestNewProfileFromReader(t *testing.T) {
	testcases := []struct {
		buf         string
		expectedErr error
	}{
		{
			// empty
			buf:         "",
			expectedErr: fmt.Errorf("EOF"),
		},
		{
			// bad json
			buf:         `{"badjson": {"ok": "asd}}`,
			expectedErr: fmt.Errorf("unexpected EOF"),
		},
		{
			// passing
			buf:         `{"config": {"path": "/some/path/"}}`,
			expectedErr: nil,
		},
	}

	buf := bytes.NewBufferString("")
	for _, testcase := range testcases {
		buf.Reset()
		buf.WriteString(testcase.buf)

		_, err := NewProfileFromReader(buf)
		if !reflect.DeepEqual(err, testcase.expectedErr) {
			t.Errorf("Expected err to be %q but got %q", testcase.expectedErr, err)
		}

		// TODO: test that read correctly.
	}
}

func TestProfile_Save(t *testing.T) {
	testcases := []struct {
		buf         string
		expectedErr error
	}{
		{
			buf:         `{"config": {"path": "/some/path/"}}`,
			expectedErr: nil,
		},
	}

	for _, testcase := range testcases {
		buf := bytes.NewBufferString(testcase.buf)
		rc, _ := NewProfileFromReader(buf)

		buf.Reset()

		err := rc.Save(buf)
		if !reflect.DeepEqual(err, testcase.expectedErr) {
			t.Errorf("Expected err to be %q but got %q", testcase.expectedErr, err)
		}

		// TODO: test that written correctly.
	}
}
