package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfile_FromJSON(t *testing.T) {
	t.Parallel()

	p, err := NewProfile("testdata/profile_valid.json")

	require.NoError(t, err)
	assert.Equal(t, "/dotfiles/root", p.DotfilesDir())
}

func TestNewProfile_FromTOML(t *testing.T) {
	t.Parallel()

	p, err := NewProfile("testdata/profile_valid.toml")

	require.NoError(t, err)
	assert.Equal(t, "/dotfiles/root", p.DotfilesDir())
}

func TestNewProfile_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/profile_invalid.json")

	assert.Error(t, err)
}

func TestNewProfile_InvalidTOML(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/profile_invalid.toml")

	assert.Error(t, err)
}

func TestNewProfile_UnknownExtension(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/somefile.badext")

	assert.Error(t, err)
}

func TestNewProfile_BadDotfilesDirectory(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/profile_bad_dotfiles.json")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an absolute path")
}

func TestNewProfile_BadSourcesDirectory(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/profile_bad_sources.json")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a relative path")
}

func TestNewProfile_BadDestinationDirectory(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/profile_bad_destination.json")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an absolute path")
}

func TestNewProfile_BadBackupDirectory(t *testing.T) {
	t.Parallel()

	_, err := NewProfile("testdata/profile_bad_backup.json")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an absolute path")
}

func TestNewProfile_ExpandEnv(t *testing.T) {
	t.Setenv("TEST_DOTFILES_DIR", "/my/dotfiles")
	t.Setenv("TEST_DESTINATION_DIR", "/my/destination")
	t.Setenv("TEST_BACKUP_DIR", "/my/backup")

	p, err := NewProfile("testdata/profile_env_vars.json")

	require.NoError(t, err)
	assert.Equal(t, "/my/dotfiles", p.DotfilesDir())
	assert.Equal(t, "/my/destination", p.DestinationDir())
	assert.Equal(t, "/my/backup", p.BackupDir())
}

func TestNewProfile_DefaultValues(t *testing.T) {
	home := os.Getenv("HOME")
	profilePath, err := filepath.Abs("testdata/profile_defaults.json")
	require.NoError(t, err)
	profileDir := filepath.Dir(profilePath)

	p, err := NewProfile(profilePath)

	require.NoError(t, err)
	assert.Equal(t, profileDir, p.DotfilesDir())
	assert.Equal(t, "", p.SourcesDir())
	assert.Equal(t, home, p.DestinationDir())
	assert.Equal(t, home+"/.dotfiles~", p.BackupDir())
}
