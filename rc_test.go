package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNewRC(t *testing.T) {
	NewRC()
}

func TestRC_SetPath(t *testing.T) {
	// set up

	configPath := "/tmp/dotbro.toml"

	// test

	rc := NewRC()

	rc.SetPath(configPath)

	if rc.Config.Path != configPath {
		t.Fatal("Fail to set configPath correctly")
	}

	// tear down
}

func TestRC_LoadNotExists(t *testing.T) {
	// set up

	RCFilepath = "/tmp/dotbro_rc.json"

	// test

	rc := NewRC()

	if err := rc.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestRC_LoadExists(t *testing.T) {
	// set up

	RCFilepath = "/tmp/dotbro_rc.json"
	configPath := "/tmp/dotbro.toml"

	setupRC := NewRC()
	setupRC.SetPath(configPath)

	f, err := os.Create(RCFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.NewEncoder(f).Encode(setupRC); err != nil {
		t.Fatal(err)
	}

	// test

	rc := NewRC()

	if err := rc.Load(); err != nil {
		t.Fatal(err)
	}

	// validate

	if rc.Config.Path != configPath {
		t.Fatal("Failed to load RC correctly")
	}

	// tear down

	os.Remove(RCFilepath)
}

func TestRC_Save(t *testing.T) {
	// set up

	RCFilepath = "/tmp/dotbro_rc.json"
	configPath := "/tmp/dotbro.toml"

	// test

	rc := NewRC()

	rc.SetPath(configPath)

	if err := rc.Save(); err != nil {
		t.Fatal(err)
	}

	// validate

	loadedRC := NewRC()

	f, err := os.Open(RCFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.NewDecoder(f).Decode(loadedRC); err != nil {
		t.Fatal(err)
	}

	if loadedRC.Config.Path != configPath {
		t.Fatal("Failed to save RC correctly")
	}

	// tear down

	os.Remove(RCFilepath)
}
