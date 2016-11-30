package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestNewConfiguration(t *testing.T) {
	testNewConfiguration_FromJSON(t)
	testNewConfiguration_FromTOML(t)

	testNewConfiguration_BadJSON(t)
	testNewConfiguration_BadTOML(t)

	testNewConfiguration_FromUnknown(t)

	testNewConfiguration_BadDotfilesDirectory(t)
	testNewConfiguration_BadSourcesDirectory(t)
}

func testNewConfiguration_FromJSON(t *testing.T) {
	// set up

	tmpFilepath := "/tmp/dotbro.json"

	file, err := os.Create(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	dirs := &Directories{
		Dotfiles: "/tmp",
	}
	conf := &Configuration{
		Directories: *dirs,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(conf); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewConfiguration(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	// tear down

	os.Remove(tmpFilepath)
}

func testNewConfiguration_FromTOML(t *testing.T) {
	// set up

	tmpFilepath := "/tmp/dotbro.toml"

	file, err := os.Create(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	dirs := &Directories{
		Dotfiles: "/tmp",
	}
	conf := &Configuration{
		Directories: *dirs,
	}

	encoder := toml.NewEncoder(file)
	if err = encoder.Encode(conf); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewConfiguration(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	// tear down

	os.Remove(tmpFilepath)
}

func testNewConfiguration_FromUnknown(t *testing.T) {
	// set up

	// test

	_, err := NewConfiguration("/tmp/somefile.badext")
	if err == nil {
		t.Fatal("Should fail because of unknown file extension")
	}

	// tear down
}

func testNewConfiguration_BadJSON(t *testing.T) {
	// set up

	tmpFilepath := "/tmp/dotbro.json"

	f, err := os.Create(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	f.WriteString("{bad json:")

	// test

	_, err = NewConfiguration(tmpFilepath)
	if err == nil {
		t.Fatal("Should error because of invalid json")
	}

	// tear down

	os.Remove(tmpFilepath)
}

func testNewConfiguration_BadTOML(t *testing.T) {
	// set up

	tmpFilepath := "/tmp/dotbro.toml"

	f, err := os.Create(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	f.WriteString("bad toml")

	// test

	_, err = NewConfiguration(tmpFilepath)
	if err == nil {
		t.Fatal("Should error because of invalid toml")
	}

	// tear down

	os.Remove(tmpFilepath)
}

func testNewConfiguration_BadDotfilesDirectory(t *testing.T) {
	// set up

	tmpFilepath := "/tmp/dotbro.json"

	file, err := os.Create(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	dirs := &Directories{
		Dotfiles: "tmp", // this path must be absolute
	}
	conf := &Configuration{
		Directories: *dirs,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(conf); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewConfiguration(tmpFilepath)
	if err == nil {
		t.Fatal("Should fail because of directories.dotfiles must be an absolute path")
	}

	// tear down

	os.Remove(tmpFilepath)
}

func testNewConfiguration_BadSourcesDirectory(t *testing.T) {
	// set up

	tmpFilepath := "/tmp/dotbro.json"

	file, err := os.Create(tmpFilepath)
	if err != nil {
		t.Fatal(err)
	}

	dirs := &Directories{
		Dotfiles: "/tmp",
		Sources:  "/dotfiles", // this path must be relative
	}
	conf := &Configuration{
		Directories: *dirs,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(conf); err != nil {
		t.Fatal(err)
	}

	// test

	_, err = NewConfiguration(tmpFilepath)
	if err == nil {
		t.Fatal("Should fail because of directories.sources must be a relative path")
	}

	// tear down

	os.Remove(tmpFilepath)
}
