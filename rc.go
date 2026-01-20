package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
)

// RCFilepath is path to dotbro's runcom file.
var RCFilepath = "${HOME}/.dotbro/profile.json"

// RC represents data for dotbro's runcom file.
type RC struct {
	Config rcConfig `json:"config"`
}

// rcConfig represents "config" section of RC data.
type rcConfig struct {
	Path  string   `json:"path,omitempty"` // Deprecated: use Paths instead
	Paths []string `json:"paths,omitempty"`
}

// NewRC returns a new RC.
func NewRC() *RC {
	return &RC{}
}

// Filepath returns the expanded path to the RC file.
func (rc *RC) Filepath() string {
	return os.ExpandEnv(RCFilepath)
}

// SetPath adds config path to the list of paths, avoiding duplicates.
func (rc *RC) SetPath(configPath string) {
	if slices.Contains(rc.Config.Paths, configPath) {
		return
	}

	rc.Config.Paths = append(rc.Config.Paths, configPath)

	// Keep old Path field in sync with the last added path for backward compatibility.
	rc.Config.Path = configPath
}

// GetPaths returns all configured paths.
func (rc *RC) GetPaths() []string {
	if len(rc.Config.Paths) > 0 {
		return rc.Config.Paths
	}

	// Backward compatibility: return Path as single-element slice
	if rc.Config.Path != "" {
		return []string{rc.Config.Path}
	}

	return nil
}

// Load reads RC data from rcFilepath.
// It maintains backward compatibility by migrating old Path field to Paths array.
func (rc *RC) Load() (err error) {
	rcFile := os.ExpandEnv(RCFilepath)

	bytes, err := ioutil.ReadFile(rcFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &rc)
	if err != nil {
		return err
	}

	// Backward compatibility: migrate old Path to Paths if needed.
	if rc.Config.Path != "" && len(rc.Config.Paths) == 0 {
		rc.Config.Paths = []string{rc.Config.Path}
	}

	return nil
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
