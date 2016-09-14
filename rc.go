package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// RCFilepath is path to dotbro's runcom file.
const RCFilepath = "${HOME}/.dotbro/profile.json"

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

// Save saves RC data to rcFilepath.
func (rc *RC) Save(configPath string) (err error) {
	rcFile := os.ExpandEnv(RCFilepath)

	if err = createPath(rcFile); err != nil {
		return err
	}

	rc.Config.Path = configPath

	bytes, err := json.MarshalIndent(rc, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(rcFile, bytes, 0666)
	return err
}
