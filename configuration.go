package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Configuration represents data from config file and additional parameters.
type Configuration struct {
	Directories directories
	Mapping     map[string]string
	Files       files
	Filepath    string
}

type directories struct {
	Dotfiles    string `toml:"dotfiles" json:"dotfiles"`
	Sources     string `toml:"sources" json:"sources"`
	Destination string
	Backup      string
}

type files struct {
	Excludes []string
}

func processConf(c Configuration) Configuration {
	c.Directories.Dotfiles = processAbsDir(
		"directories.dotfiles_root",
		c.Directories.Dotfiles,
		path.Dir(c.Filepath),
	)

	c.Directories.Sources = processRelDir(
		"directories.dotfiles_source",
		c.Directories.Sources,
		"",
	)

	c.Directories.Destination = processAbsDir(
		"directories.destination",
		c.Directories.Destination,
		os.Getenv("HOME"),
	)

	c.Directories.Backup = processAbsDir(
		"directories.backup",
		c.Directories.Backup,
		os.Getenv("HOME")+"/.dotfiles~",
	)

	return c
}

func processAbsDir(name string, dir string, defaultValue string) string {
	if dir == "" {
		return defaultValue
	}

	// parse possible ENV variable
	dir = os.ExpandEnv(dir)

	if !path.IsAbs(dir) {
		outError("'%s' must be an absolute path.", name)
		exit(1)
	}

	return dir
}

func processRelDir(name string, dir string, defaultValue string) string {
	if dir == "" {
		return defaultValue
	}

	// parse possible ENV variable
	dir = os.ExpandEnv(dir)

	if path.IsAbs(dir) {
		outError("'%s' must be relative path (to 'directories.dotfiles_root').\n", name)
		exit(1)
	}

	return dir
}

func fromTOML(filename string) (conf Configuration, err error) {
	logger.msg("Parsing config file %s", filename)

	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		return conf, err
	}

	conf.Filepath = filename

	conf = processConf(conf)

	return conf, nil
}

func fromJSON(filename string) (conf Configuration, err error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return conf, err
	}

	logger.msg("Parsing config file %s", filename)

	json.Unmarshal(file, &conf)

	conf.Filepath = filename

	conf = processConf(conf)

	return conf, nil
}

func configurationFromFile(filename string) (conf Configuration, err error) {
	switch filepath.Ext(filename) {
	case ".toml":
		return fromTOML(filename)
	case ".json":
		return fromJSON(filename)
	default:
		err = fmt.Errorf("Cannot read config file %s : unknown extension. Supported: conf, toml.", filename)
		return conf, err
	}
}
