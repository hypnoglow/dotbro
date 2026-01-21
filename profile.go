package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// Profile represents data from a profile file and additional parameters.
type Profile struct {
	Directories Directories
	Mapping     map[string]string
	Files       Files
	Filepath    string
}

// Directories represents [directories] section of a profile.
type Directories struct {
	Dotfiles    string `toml:"dotfiles" json:"dotfiles"`
	Sources     string `toml:"sources" json:"sources"`
	Destination string `toml:"destination" json:"destination"`
	Backup      string `toml:"backup" json:"backup"`
}

// Files represents [files] section of a profile.
type Files struct {
	Excludes []string
}

// NewProfile returns a new Profile.
func NewProfile(filename string) (p *Profile, err error) {
	switch filepath.Ext(filename) {
	case ".toml":
		p, err = profileFromTOML(filename)
	case ".json":
		p, err = profileFromJSON(filename)
	default:
		err = fmt.Errorf("unknown profile file extension %s: supported extensions are .toml and .json", filename)
	}

	if err != nil {
		return nil, err
	}

	p, err = processProfile(p)
	if err != nil {
		return nil, err
	}

	p.Filepath = filename
	return p, nil
}

func profileFromTOML(filename string) (p *Profile, err error) {
	if _, err = toml.DecodeFile(filename, &p); err != nil {
		return nil, err
	}

	return p, nil
}

func profileFromJSON(filename string) (p *Profile, err error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(file, &p); err != nil {
		return nil, err
	}

	return p, nil
}

func processProfile(p *Profile) (*Profile, error) {
	params := getDirectoriesParams(p.Filepath)

	t := reflect.TypeOf(p.Directories)
	r := reflect.ValueOf(&p.Directories).Elem()

	for i := 0; i < r.NumField(); i++ {
		name := t.Field(i).Name
		field := r.FieldByName(name)

		value := os.ExpandEnv(field.String())
		if value == "" {
			value = params[name].defaultValue
		}

		err := checkIfDirCorrect(name, value, params[name].isRelative)
		if err != nil {
			return nil, err
		}

		field.SetString(value)
	}

	return p, nil
}

type directoryParam struct {
	defaultValue string
	isRelative   bool
}

func getDirectoriesParams(profilePath string) map[string]directoryParam {
	params := map[string]directoryParam{
		"Dotfiles": {
			defaultValue: path.Dir(profilePath),
			isRelative:   false,
		},
		"Sources": {
			defaultValue: "",
			isRelative:   true,
		},
		"Destination": {
			defaultValue: os.Getenv("HOME"),
			isRelative:   false,
		},
		"Backup": {
			defaultValue: os.Getenv("HOME") + "/.dotfiles~",
			isRelative:   false,
		},
	}
	return params
}

func checkIfDirCorrect(fieldName, dir string, isRelative bool) error {
	if !isRelative && !path.IsAbs(dir) {
		return fmt.Errorf(
			"'directories.%s' must be an absolute path",
			strings.ToLower(fieldName),
		)
	}

	if isRelative && path.IsAbs(dir) {
		return fmt.Errorf(
			"'directories.%s' must be a relative path (to 'directories.dotfiles_root').\n",
			strings.ToLower(fieldName),
		)
	}

	return nil
}
