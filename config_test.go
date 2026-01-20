package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
)

const testConfigPath = "/tmp/dotbro_config.json"
const testLegacyConfigPath = "/tmp/dotbro_legacy_profile.json"

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestConfig_SetPath(t *testing.T) {
	cfg := NewConfig(testLogger())

	cfg.AddProfile(testTOMLConfigPath)

	if len(cfg.Profiles) != 1 || cfg.Profiles[0].Path != testTOMLConfigPath {
		t.Fatal("Fail to set configPath correctly")
	}
}

func TestConfig_SetPath_NoDuplicates(t *testing.T) {
	cfg := NewConfig(testLogger())

	cfg.AddProfile(testTOMLConfigPath)
	cfg.AddProfile(testTOMLConfigPath)

	if len(cfg.Profiles) != 1 {
		t.Fatal("SetPath should not add duplicates")
	}
}

func TestConfig_GetPaths(t *testing.T) {
	cfg := NewConfig(testLogger())
	cfg.AddProfile("/path/one")
	cfg.AddProfile("/path/two")

	paths := cfg.GetProfilePaths()

	if len(paths) != 2 || paths[0] != "/path/one" || paths[1] != "/path/two" {
		t.Fatal("GetPaths returned incorrect paths")
	}
}

func TestConfig_LoadNotExists(t *testing.T) {
	// set up
	configFilepath = testConfigPath
	legacyConfigFilepath = testLegacyConfigPath

	// ensure files don't exist
	os.Remove(testConfigPath)
	os.Remove(testLegacyConfigPath)

	// test
	cfg := NewConfig(testLogger())
	ctx := context.Background()

	if err := cfg.Load(ctx); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Profiles) != 0 {
		t.Fatal("Expected empty profiles when no config exists")
	}
}

func TestConfig_LoadExists(t *testing.T) {
	// set up
	configFilepath = testConfigPath
	legacyConfigFilepath = testLegacyConfigPath

	// ensure legacy file doesn't exist
	os.Remove(testLegacyConfigPath)

	setupCfg := NewConfig(testLogger())
	setupCfg.AddProfile(testTOMLConfigPath)

	f, err := os.Create(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.NewEncoder(f).Encode(setupCfg); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// test
	cfg := NewConfig(testLogger())
	ctx := context.Background()

	if err := cfg.Load(ctx); err != nil {
		t.Fatal(err)
	}

	// validate
	if len(cfg.Profiles) != 1 || cfg.Profiles[0].Path != testTOMLConfigPath {
		t.Fatal("Failed to load Config correctly")
	}

	// tear down
	os.Remove(configFilepath)
}

func TestConfig_Save(t *testing.T) {
	// set up
	configFilepath = testConfigPath

	// test
	cfg := NewConfig(testLogger())
	ctx := context.Background()

	cfg.AddProfile(testTOMLConfigPath)

	if err := cfg.Save(ctx); err != nil {
		t.Fatal(err)
	}

	// validate
	loadedCfg := NewConfig(testLogger())

	f, err := os.Open(configFilepath)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.NewDecoder(f).Decode(loadedCfg); err != nil {
		t.Fatal(err)
	}
	f.Close()

	if len(loadedCfg.Profiles) != 1 || loadedCfg.Profiles[0].Path != testTOMLConfigPath {
		t.Fatal("Failed to save Config correctly")
	}

	// tear down
	os.Remove(configFilepath)
}

func TestConfig_LoadMigration(t *testing.T) {
	// set up - create legacy profile.json
	configFilepath = testConfigPath
	legacyConfigFilepath = testLegacyConfigPath

	// Clean up any existing files
	os.Remove(testConfigPath)
	os.Remove(testLegacyConfigPath)

	// Create legacy format file
	legacy := legacyRC{
		Config: legacyRCConfig{
			Path:  testTOMLConfigPath,
			Paths: []string{testTOMLConfigPath, "/another/path"},
		},
	}

	f, err := os.Create(testLegacyConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.NewEncoder(f).Encode(legacy); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// test - Load should migrate legacy file to new location
	cfg := NewConfig(testLogger())
	ctx := context.Background()

	if err := cfg.Load(ctx); err != nil {
		t.Fatal(err)
	}

	// validate - config.json should exist, profile.json should not
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Fatal("config.json should exist after migration")
	}

	if _, err := os.Stat(testLegacyConfigPath); !os.IsNotExist(err) {
		t.Fatal("profile.json should not exist after migration")
	}

	if len(cfg.Profiles) != 2 {
		t.Fatalf("Expected 2 profiles after migration, got %d", len(cfg.Profiles))
	}

	if cfg.Profiles[0].Path != testTOMLConfigPath {
		t.Fatal("Failed to load migrated Config correctly")
	}

	if cfg.Profiles[1].Path != "/another/path" {
		t.Fatal("Failed to load second path from migrated Config")
	}

	// tear down
	os.Remove(testConfigPath)
}

func TestConfig_LoadMigrationFromSinglePath(t *testing.T) {
	// set up - create legacy profile.json with only Path field (old format)
	configFilepath = testConfigPath
	legacyConfigFilepath = testLegacyConfigPath

	// Clean up any existing files
	os.Remove(testConfigPath)
	os.Remove(testLegacyConfigPath)

	// Create legacy format file with only Path field (no Paths array)
	legacy := legacyRC{
		Config: legacyRCConfig{
			Path: testTOMLConfigPath,
		},
	}

	f, err := os.Create(testLegacyConfigPath)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.NewEncoder(f).Encode(legacy); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// test - Load should migrate legacy file
	cfg := NewConfig(testLogger())
	ctx := context.Background()

	if err := cfg.Load(ctx); err != nil {
		t.Fatal(err)
	}

	// validate
	if len(cfg.Profiles) != 1 {
		t.Fatalf("Expected 1 profile after migration from single Path, got %d", len(cfg.Profiles))
	}

	if cfg.Profiles[0].Path != testTOMLConfigPath {
		t.Fatal("Failed to load migrated Config correctly")
	}

	// tear down
	os.Remove(testConfigPath)
}
