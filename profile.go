package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Profile represents a loaded profile with its file path and data.
type Profile struct {
	// filepath is the path to the profile file itself.
	filepath string

	// data contains the parsed profile data.
	data ProfileData
}

// ProfileData represents the data structure of a profile file.
type ProfileData struct {
	// Directories contains paths configuration for dotfiles management.
	Directories Directories `toml:"directories" json:"directories"`

	// Mapping defines source-to-destination file mappings.
	Mapping map[string]string `toml:"mapping" json:"mapping"`

	// Files contains file filtering options.
	Files Files `toml:"files" json:"files"`
}

// Directories represents [directories] section of a profile.
type Directories struct {
	// Dotfiles is the root directory containing dotfiles to manage.
	Dotfiles string `toml:"dotfiles" json:"dotfiles"`

	// Sources is a subdirectory within Dotfiles containing actual dotfiles.
	// Deprecated: use Dotfiles directly.
	Sources string `toml:"sources" json:"sources"`

	// Destination is the target directory where symlinks will be created.
	Destination string `toml:"destination" json:"destination"`

	// Backup is the directory for storing original files before symlinking.
	Backup string `toml:"backup" json:"backup"`
}

// Files represents [files] section of a profile.
type Files struct {
	Excludes []string
}

// NewProfile returns a new Profile.
func NewProfile(filename string) (*Profile, error) {
	var data ProfileData
	var err error

	switch filepath.Ext(filename) {
	case ".toml":
		data, err = profileDataFromTOML(filename)
	case ".json":
		data, err = profileDataFromJSON(filename)
	default:
		err = fmt.Errorf("unknown profile file extension %s: supported extensions are .toml and .json", filename)
	}

	if err != nil {
		return nil, err
	}

	data, err = processProfileData(data, filename)
	if err != nil {
		return nil, err
	}

	return &Profile{
		filepath: filename,
		data:     data,
	}, nil
}

// Filepath returns the path to the profile file.
func (p Profile) Filepath() string {
	return p.filepath
}

// Data returns the profile data.
func (p Profile) Data() ProfileData {
	return p.data
}

// DotfilesDir returns the dotfiles directory path.
func (p Profile) DotfilesDir() string {
	return p.data.Directories.Dotfiles
}

// SourcesDir returns the sources subdirectory path.
// Deprecated: use DotfilesDir directly.
func (p Profile) SourcesDir() string {
	return p.data.Directories.Sources
}

// DestinationDir returns the destination directory path.
func (p Profile) DestinationDir() string {
	return p.data.Directories.Destination
}

// BackupDir returns the backup directory path.
func (p Profile) BackupDir() string {
	return p.data.Directories.Backup
}

func profileDataFromTOML(filename string) (ProfileData, error) {
	var data ProfileData
	if _, err := toml.DecodeFile(filename, &data); err != nil {
		return ProfileData{}, err
	}
	return data, nil
}

func profileDataFromJSON(filename string) (ProfileData, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return ProfileData{}, err
	}

	var data ProfileData
	if err = json.Unmarshal(file, &data); err != nil {
		return ProfileData{}, err
	}
	return data, nil
}

func processProfileData(data ProfileData, profilePath string) (ProfileData, error) {
	home := os.Getenv("HOME")
	profileDir := path.Dir(profilePath)

	dirs := [4]directory{
		{
			name:         "dotfiles",
			value:        data.Directories.Dotfiles,
			defaultValue: profileDir,
			relative:     false,
		},
		{
			name:         "sources",
			value:        data.Directories.Sources,
			defaultValue: "",
			relative:     true,
		},
		{
			name:         "destination",
			value:        data.Directories.Destination,
			defaultValue: home,
			relative:     false,
		},
		{
			name:         "backup",
			value:        data.Directories.Backup,
			defaultValue: home + "/.dotfiles~",
			relative:     false,
		},
	}

	for i := range dirs {
		dirs[i].value = os.ExpandEnv(dirs[i].value)
		if dirs[i].value == "" {
			dirs[i].value = dirs[i].defaultValue
		}

		if dirs[i].relative {
			if err := checkDirectoryRelative(dirs[i].name, dirs[i].value); err != nil {
				return ProfileData{}, err
			}
		} else {
			if err := checkDirectoryAbsolute(dirs[i].name, dirs[i].value); err != nil {
				return ProfileData{}, err
			}
		}
	}

	data.Directories.Dotfiles = dirs[0].value
	data.Directories.Sources = dirs[1].value
	data.Directories.Destination = dirs[2].value
	data.Directories.Backup = dirs[3].value

	return data, nil
}

type directory struct {
	name         string
	value        string
	defaultValue string
	relative     bool
}

func checkDirectoryAbsolute(name, dir string) error {
	if !path.IsAbs(dir) {
		return fmt.Errorf(
			"'directories.%s' must be an absolute path",
			name,
		)
	}
	return nil
}

func checkDirectoryRelative(name, dir string) error {
	if path.IsAbs(dir) {
		return fmt.Errorf(
			"'directories.%s' must be a relative path",
			name,
		)
	}
	return nil
}
