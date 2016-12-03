package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

// Configuration represents data from config file and additional parameters.
type Configuration struct {
	Directories Directories
	Mapping     map[string]string
	Files       Files
	Filepath    string
}

// Directories represents [directories] section of a config.
type Directories struct {
	Dotfiles    string `toml:"dotfiles" json:"dotfiles"`
	Sources     string `toml:"sources" json:"sources"`
	Destination string `toml:"destination" json:"destination"`
	Backup      string `toml:"backup" json:"backup"`
}

// Files represents [files] section of a config.
type Files struct {
	Excludes []string
}

// NewConfiguration returns a new Configuration.
func NewConfiguration(filename string) (conf *Configuration, err error) {
	switch filepath.Ext(filename) {
	case ".toml":
		conf, err = fromTOML(filename)
	case ".json":
		conf, err = fromJSON(filename)
	default:
		err = fmt.Errorf("Cannot read config file %s : unknown extension. Supported: conf, toml.", filename)
	}

	if err != nil {
		return nil, err
	}

	conf, err = processConf(conf)
	if err != nil {
		return nil, err
	}

	conf.Filepath = filename
	return conf, nil
}

func fromTOML(filename string) (conf *Configuration, err error) {
	if _, err = toml.DecodeFile(filename, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func fromJSON(filename string) (conf *Configuration, err error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(file, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func processConf(c *Configuration) (*Configuration, error) {
	params := getDirectoriesParams(c.Filepath)

	t := reflect.TypeOf(c.Directories)
	r := reflect.ValueOf(&c.Directories).Elem()

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

	return c, nil
}

type directoryParam struct {
	defaultValue string
	isRelative   bool
}

func getDirectoriesParams(configPath string) map[string]directoryParam {
	params := map[string]directoryParam{
		"Dotfiles": directoryParam{
			defaultValue: path.Dir(configPath),
			isRelative:   false,
		},
		"Sources": directoryParam{
			defaultValue: "",
			isRelative:   true,
		},
		"Destination": directoryParam{
			defaultValue: os.Getenv("HOME"),
			isRelative:   false,
		},
		"Backup": directoryParam{
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
