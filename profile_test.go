package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

const (
	testJSONProfilePath = "/tmp/dotbro.json"
	testTOMLProfilePath = "/tmp/dotbro.toml"
)

func TestNewProfile(t *testing.T) {
	testNewProfile_FromJSON(t)
	testNewProfile_FromTOML(t)

	testNewProfile_BadJSON(t)
	testNewProfile_BadTOML(t)

	testNewProfile_FromUnknown(t)

	testNewProfile_BadDotfilesDirectory(t)
	testNewProfile_BadSourcesDirectory(t)
}

func testNewProfile_FromJSON(t *testing.T) {
	// set up

	file, err := os.Create(testJSONProfilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONProfilePath)

	dirs := &Directories{
		Dotfiles: "/tmp",
	}
	p := &Profile{
		Directories: *dirs,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(p); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewProfile(testJSONProfilePath)
	if err != nil {
		t.Fatal(err)
	}
}

func testNewProfile_FromTOML(t *testing.T) {
	// set up

	file, err := os.Create(testTOMLProfilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testTOMLProfilePath)

	dirs := &Directories{
		Dotfiles: "/tmp",
	}
	p := &Profile{
		Directories: *dirs,
	}

	encoder := toml.NewEncoder(file)
	if err = encoder.Encode(p); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewProfile(testTOMLProfilePath)
	if err != nil {
		t.Fatal(err)
	}
}

func testNewProfile_FromUnknown(t *testing.T) {
	// set up

	// test

	_, err := NewProfile("/tmp/somefile.badext")
	if err == nil {
		t.Fatal("Should fail because of unknown file extension")
	}

	// tear down
}

func testNewProfile_BadJSON(t *testing.T) {
	// set up

	f, err := os.Create(testJSONProfilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONProfilePath)

	_, _ = f.WriteString("{bad json:")

	// test

	_, err = NewProfile(testJSONProfilePath)
	if err == nil {
		t.Fatal("Should error because of invalid json")
	}
}

func testNewProfile_BadTOML(t *testing.T) {
	// set up

	f, err := os.Create(testTOMLProfilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testTOMLProfilePath)

	_, _ = f.WriteString("bad toml")

	// test

	_, err = NewProfile(testTOMLProfilePath)
	if err == nil {
		t.Fatal("Should error because of invalid toml")
	}
}

func testNewProfile_BadDotfilesDirectory(t *testing.T) {
	// set up

	file, err := os.Create(testJSONProfilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONProfilePath)

	dirs := &Directories{
		Dotfiles: "tmp", // this path must be absolute
	}
	p := &Profile{
		Directories: *dirs,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(p); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewProfile(testJSONProfilePath)
	if err == nil {
		t.Fatal("Should fail because of directories.dotfiles must be an absolute path")
	}
}

func testNewProfile_BadSourcesDirectory(t *testing.T) {
	// set up

	file, err := os.Create(testJSONProfilePath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONProfilePath)

	dirs := &Directories{
		Dotfiles: "/tmp",
		Sources:  "/dotfiles", // this path must be relative
	}
	p := &Profile{
		Directories: *dirs,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(p); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewProfile(testJSONProfilePath)
	if err == nil {
		t.Fatal("Should fail because of directories.sources must be a relative path")
	}
}
