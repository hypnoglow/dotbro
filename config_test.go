package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_AddProfile(t *testing.T) {
	t.Parallel()

	conf := NewConfig(newDiscardLogger(), "", "")

	conf.AddProfile("/some/path")

	require.Len(t, conf.data.Profiles, 1)
	assert.Equal(t, "/some/path", conf.data.Profiles[0].Path)
}

func TestConfig_AddProfile_NoDuplicates(t *testing.T) {
	t.Parallel()

	conf := NewConfig(newDiscardLogger(), "", "")

	conf.AddProfile("/some/path")
	conf.AddProfile("/some/path")

	assert.Len(t, conf.data.Profiles, 1)
}

func TestConfig_GetProfilePaths(t *testing.T) {
	t.Parallel()

	conf := NewConfig(newDiscardLogger(), "", "")
	conf.AddProfile("/path/one")
	conf.AddProfile("/path/two")

	paths := conf.GetProfilePaths()

	require.Len(t, paths, 2)
	assert.Equal(t, "/path/one", paths[0])
	assert.Equal(t, "/path/two", paths[1])
}

func TestConfig_Load_NotExists(t *testing.T) {
	t.Parallel()

	conf := NewConfig(
		newDiscardLogger(),
		"testdata/non_existent_config.json",
		"testdata/non_existent_profile.json",
	)

	require.NoError(t, conf.Load(t.Context()))
	assert.Empty(t, conf.data.Profiles)
}

func TestConfig_Load_Exists(t *testing.T) {
	t.Parallel()

	conf := NewConfig(
		newDiscardLogger(),
		"testdata/config.json",
		"testdata/non_existent_profile.json",
	)

	require.NoError(t, conf.Load(t.Context()))

	require.Len(t, conf.data.Profiles, 1)
	assert.Equal(t, "/test/profile/path", conf.data.Profiles[0].Path)
}

func TestConfig_Load_MigrationPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	legacyPath := filepath.Join(tmpDir, "profile.json")

	copyFile(t, "testdata/legacy_profile_1.json", legacyPath)

	conf := NewConfig(newDiscardLogger(), configPath, legacyPath)

	require.NoError(t, conf.Load(t.Context()))

	// Validate - config.json should exist, profile.json should not
	assert.FileExists(t, configPath)
	assert.NoFileExists(t, legacyPath)

	require.Len(t, conf.data.Profiles, 2)
	assert.Equal(t, "/first/path", conf.data.Profiles[0].Path)
	assert.Equal(t, "/second/path", conf.data.Profiles[1].Path)
}

func TestConfig_Load_MigrationPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	legacyPath := filepath.Join(tmpDir, "profile.json")

	copyFile(t, "testdata/legacy_profile_2.json", legacyPath)

	conf := NewConfig(newDiscardLogger(), configPath, legacyPath)

	require.NoError(t, conf.Load(t.Context()))

	// Validate - config.json should exist, profile.json should not
	assert.FileExists(t, configPath)
	assert.NoFileExists(t, legacyPath)

	require.Len(t, conf.data.Profiles, 1)
	assert.Equal(t, "/one/path", conf.data.Profiles[0].Path)
}

func TestConfig_Save(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	conf := NewConfig(newDiscardLogger(), configPath, "")

	conf.AddProfile("/test/profile/path")

	require.NoError(t, conf.Save(t.Context()))

	// Validate

	var data ConfigData

	f, err := os.Open(configPath)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(f).Decode(&data))
	f.Close()

	require.Len(t, data.Profiles, 1)
	assert.Equal(t, "/test/profile/path", data.Profiles[0].Path)
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()

	srcFile, err := os.Open(src)
	require.NoError(t, err)
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	require.NoError(t, err)
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	require.NoError(t, err)
}
