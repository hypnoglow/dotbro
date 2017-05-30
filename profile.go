package main

import (
	"encoding/json"
	"io"
)

// ProfilePath is path to dotbro profile file.
const ProfilePath = "${HOME}/.dotbro/profile.json"

// Profile represents dotbro profile.
type Profile struct {
	Config profileConfig `json:"config"`
}

// profileConfig represents "config" section of Profile.
type profileConfig struct {
	Path string `json:"path"`
}

// NewProfile returns a new Profile.
func NewProfile() *Profile {
	return &Profile{}
}

// NewRCFromFile returns a new Profile which is read from the reader.
func NewProfileFromReader(reader io.Reader) (rc Profile, err error) {
	dec := json.NewDecoder(reader)
	err = dec.Decode(&rc)
	return rc, err
}

// Save saves the Profile data to the writer..
func (rc Profile) Save(writer io.Writer) (err error) {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "    ")
	return enc.Encode(rc)
}
