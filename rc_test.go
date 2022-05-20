package main

import (
	"encoding/json"
	"os"
	"testing"
)

const testRCPath = "/tmp/dotbro_rc.json"

func TestRC_SetPath(t *testing.T) {
	rc := NewRC()

	rc.SetPath(testTOMLConfigPath)

	if rc.Config.Path != testTOMLConfigPath {
		t.Fatal("Fail to set configPath correctly")
	}
}

func TestRC_LoadNotExists(t *testing.T) {
	// set up

	RCFilepath = testRCPath

	// test

	rc := NewRC()

	if err := rc.Load(); err != nil {
		t.Fatal(err)
	}
}

func TestRC_LoadExists(t *testing.T) {
	// set up

	RCFilepath = testRCPath

	setupRC := NewRC()
	setupRC.SetPath(testTOMLConfigPath)

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

	if rc.Config.Path != testTOMLConfigPath {
		t.Fatal("Failed to load RC correctly")
	}

	// tear down

	os.Remove(RCFilepath)
}

func TestRC_Save(t *testing.T) {
	// set up

	RCFilepath = testRCPath

	// test

	rc := NewRC()

	rc.SetPath(testTOMLConfigPath)

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

	if loadedRC.Config.Path != testTOMLConfigPath {
		t.Fatal("Failed to save RC correctly")
	}

	// tear down

	os.Remove(RCFilepath)
}
