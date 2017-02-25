package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// RCFilepath is path to dotbro's runcom file.
var RCFilepath = "${HOME}/.dotbro/profile.json"

// RC represents data for dotbro's runcom file.
type RC struct {
	Config rcConfig `json:"config"`
}

// rcConfig represents "config" section of RC data
type rcConfig struct {
	Path string `json:"path"`
}

// NewRC returns a new RC.
func NewRC() *RC {
	return &RC{}
}

// SetPath sets config path.
func (rc *RC) SetPath(configPath string) {
	rc.Config.Path = configPath
}

// Load reads RC data from rcFilepath.
func (rc *RC) Load() (err error) {
	rcFile := os.ExpandEnv(RCFilepath)

	bytes, err := ioutil.ReadFile(rcFile)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &rc)
	return err
}

// Save saves RC data to the file located at `RCFilepath`.
func (rc *RC) Save() (err error) {
	rcFile := os.ExpandEnv(RCFilepath)

	if err = osfs.MkdirAll(filepath.Dir(rcFile), 0700); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(rc, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(rcFile, bytes, 0666)
	return err
}
