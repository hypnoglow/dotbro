package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const rcFilepath = "${HOME}/.dotbro/profile.json"

// RC represents data for dotbro's runcom file.
type RC struct {
	Config rcConfig `json:"config"`
}

// rcConfig represents "config" section of RC data
type rcConfig struct {
	Path string `json:"path"`
}

// readRC reads RC from rcFilepath and returns it.
func readRC() (rc RC, err error) {
	rcFile := os.ExpandEnv(rcFilepath)

	bytes, err := ioutil.ReadFile(rcFile)
	if os.IsNotExist(err) {
		return rc, nil
	} else if err != nil {
		return rc, err
	}

	err = json.Unmarshal(bytes, &rc)
	return rc, err
}

// saveRC creates RC, saves it to rcFilepath and returns it.
func saveRC(configPath string) (rc RC, err error) {
	rcFile := os.ExpandEnv(rcFilepath)

	err = createPath(rcFile)
	if err != nil {
		return rc, err
	}

	rc = RC{
		Config: rcConfig{
			Path: configPath,
		},
	}

	bytes, err := json.MarshalIndent(rc, "", "    ")
	if err != nil {
		return rc, err
	}

	ioutil.WriteFile(rcFile, bytes, 0666)
	return rc, err
}
