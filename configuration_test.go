package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/BurntSushi/toml"
)

const (
	testJSONConfigPath = "/tmp/dotbro.json"
	testTOMLConfigPath = "/tmp/dotbro.toml"
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

	file, err := os.Create(testJSONConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONConfigPath)

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

	_, err = NewConfiguration(testJSONConfigPath)
	if err != nil {
		t.Fatal(err)
	}
}

func testNewConfiguration_FromTOML(t *testing.T) {
	// set up

	file, err := os.Create(testTOMLConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testTOMLConfigPath)

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

	_, err = NewConfiguration(testTOMLConfigPath)
	if err != nil {
		t.Fatal(err)
	}
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

	f, err := os.Create(testJSONConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONConfigPath)

	_, _ = f.WriteString("{bad json:")

	// test

	_, err = NewConfiguration(testJSONConfigPath)
	if err == nil {
		t.Fatal("Should error because of invalid json")
	}
}

func testNewConfiguration_BadTOML(t *testing.T) {
	// set up

	f, err := os.Create(testTOMLConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testTOMLConfigPath)

	_, _ = f.WriteString("bad toml")

	// test

	_, err = NewConfiguration(testTOMLConfigPath)
	if err == nil {
		t.Fatal("Should error because of invalid toml")
	}
}

func testNewConfiguration_BadDotfilesDirectory(t *testing.T) {
	// set up

	file, err := os.Create(testJSONConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONConfigPath)

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

	_, err = NewConfiguration(testJSONConfigPath)
	if err == nil {
		t.Fatal("Should fail because of directories.dotfiles must be an absolute path")
	}
}

func testNewConfiguration_BadSourcesDirectory(t *testing.T) {
	// set up

	file, err := os.Create(testJSONConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(testJSONConfigPath)

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

	_, err = NewConfiguration(testJSONConfigPath)
	if err == nil {
		t.Fatal("Should fail because of directories.sources must be a relative path")
	}
}
