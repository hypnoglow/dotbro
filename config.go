package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// configFilepath is path to dotbro config file.
var configFilepath = "${HOME}/.dotbro/config.json"

// legacyConfigFilepath is the old config file path for backward compatibility migration.
var legacyConfigFilepath = "${HOME}/.dotbro/profile.json"

// Config represents dotbro config.
type Config struct {
	Profiles []Profile `json:"profiles"`

	logger *slog.Logger `json:"-"`
}

// Profile represents a single profile entry.
type Profile struct {
	Path string `json:"path"`
}

// legacyRC represents old profile.json format for migration.
type legacyRC struct {
	Config legacyRCConfig `json:"config"`
}

type legacyRCConfig struct {
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

// NewConfig returns a new Config.
func NewConfig(logger *slog.Logger) *Config {
	return &Config{logger: logger}
}

// AddProfile adds a profile to the list, avoiding duplicates.
func (c *Config) AddProfile(path string) {
	for _, p := range c.Profiles {
		if p.Path == path {
			return
		}
	}
	c.Profiles = append(c.Profiles, Profile{Path: path})
}

// GetProfilePaths returns all configured profile paths.
func (c *Config) GetProfilePaths() []string {
	paths := make([]string, 0, len(c.Profiles))
	for _, p := range c.Profiles {
		paths = append(paths, p.Path)
	}
	return paths
}

// Load reads Config data from configFilepath.
// It maintains backward compatibility by migrating old profile.json format.
func (c *Config) Load(ctx context.Context) error {
	configFile := os.ExpandEnv(configFilepath)
	legacyFile := os.ExpandEnv(legacyConfigFilepath)

	// Try to load new format first
	data, err := os.ReadFile(configFile)
	if err == nil {
		c.logger.DebugContext(ctx, "Loaded config", slog.String("path", configFile))
		return json.Unmarshal(data, c)
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("read config file: %w", err)
	}

	// config.json doesn't exist, try to migrate from legacy profile.json
	legacyData, err := os.ReadFile(legacyFile)
	if os.IsNotExist(err) {
		// Neither file exists, nothing to load
		c.logger.DebugContext(ctx, "No config file found, starting fresh")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read legacy config file: %w", err)
	}

	c.logger.InfoContext(ctx, "Migrating legacy config",
		slog.String("from", legacyFile),
		slog.String("to", configFile))

	// Parse legacy format
	var legacy legacyRC
	if err := json.Unmarshal(legacyData, &legacy); err != nil {
		return fmt.Errorf("parse legacy config file: %w", err)
	}

	// Convert to new format
	c.migrateFromLegacy(&legacy)

	// Save in new format
	if err := c.Save(ctx); err != nil {
		return fmt.Errorf("save migrated config: %w", err)
	}

	// Remove old file
	if err := os.Remove(legacyFile); err != nil {
		return fmt.Errorf("remove legacy config file: %w", err)
	}

	c.logger.InfoContext(ctx, "Config migration completed")

	return nil
}

// Save saves Config data to the file located at configFilepath.
func (c *Config) Save(ctx context.Context) error {
	configFile := os.ExpandEnv(configFilepath)

	if err := osfs.MkdirAll(filepath.Dir(configFile), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return err
	}

	c.logger.DebugContext(ctx, "Saved config", slog.String("path", configFile))

	return nil
}

// migrateFromLegacy converts legacy RC format to new Config format.
func (c *Config) migrateFromLegacy(legacy *legacyRC) {
	paths := legacy.Config.Paths
	if len(paths) == 0 && legacy.Config.Path != "" {
		paths = []string{legacy.Config.Path}
	}

	for _, p := range paths {
		c.Profiles = append(c.Profiles, Profile{Path: p})
	}
}
