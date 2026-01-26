package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// defaultConfigFilepath is path to dotbro config file.
const defaultConfigFilepath = "${HOME}/.dotbro/config.json"

// defaultLegacyConfigFilepath is the old config file path for backward compatibility migration.
const defaultLegacyConfigFilepath = "${HOME}/.dotbro/profile.json"

// Config represents dotbro config.
type Config struct {
	logger           *slog.Logger
	configPath       string
	legacyConfigPath string
	data             ConfigData
}

// ConfigData represents the JSON representation of the config file.
type ConfigData struct {
	Profiles []ConfigProfile `json:"profiles"`
}

// ConfigProfile represents a single profile entry in the config.
type ConfigProfile struct {
	Path string `json:"path"`
}

// legacyConfigData represents old profile.json format for migration.
type legacyConfigData struct {
	Config legacyConfigDataConfig `json:"config"`
}

type legacyConfigDataConfig struct {
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths,omitempty"`
}

// NewConfig returns a new Config.
func NewConfig(logger *slog.Logger, configPath, legacyConfigPath string) *Config {
	return &Config{
		logger:           logger,
		configPath:       os.ExpandEnv(configPath),
		legacyConfigPath: os.ExpandEnv(legacyConfigPath),
	}
}

// AddProfile adds a profile to the list, avoiding duplicates.
func (c *Config) AddProfile(path string) {
	for _, p := range c.data.Profiles {
		if p.Path == path {
			return
		}
	}
	c.data.Profiles = append(c.data.Profiles, ConfigProfile{Path: path})
}

// GetProfilePaths returns all configured profile paths.
func (c *Config) GetProfilePaths() []string {
	paths := make([]string, 0, len(c.data.Profiles))
	for _, p := range c.data.Profiles {
		paths = append(paths, p.Path)
	}
	return paths
}

// Load reads Config data from config file.
// It maintains backward compatibility by migrating old profile.json format.
func (c *Config) Load(ctx context.Context) error {
	// Try to load new format first
	data, err := os.ReadFile(c.configPath)
	if err == nil {
		c.logger.DebugContext(ctx, "Loaded config", slog.String("path", c.configPath))
		return json.Unmarshal(data, &c.data)
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("read config file: %w", err)
	}

	// config.json doesn't exist, try to migrate from legacy profile.json
	legacyData, err := os.ReadFile(c.legacyConfigPath)
	if os.IsNotExist(err) {
		// Neither file exists, nothing to load
		c.logger.DebugContext(ctx, "No config file found, starting fresh")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read legacy config file: %w", err)
	}

	c.logger.InfoContext(ctx, "Migrating legacy config",
		slog.String("from", c.legacyConfigPath),
		slog.String("to", c.configPath))

	// Parse legacy format
	var legacy legacyConfigData
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
	if err := os.Remove(c.legacyConfigPath); err != nil {
		return fmt.Errorf("remove legacy config file: %w", err)
	}

	c.logger.InfoContext(ctx, "Config migration completed")

	return nil
}

// Save saves Config data to the config file.
func (c *Config) Save(ctx context.Context) error {
	if err := osfs.MkdirAll(filepath.Dir(c.configPath), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.data, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(c.configPath, data, 0600); err != nil {
		return err
	}

	c.logger.DebugContext(ctx, "Saved config", slog.String("path", c.configPath))

	return nil
}

// migrateFromLegacy converts legacy RC format to new Config format.
func (c *Config) migrateFromLegacy(data *legacyConfigData) {
	paths := data.Config.Paths
	if len(paths) == 0 && data.Config.Path != "" {
		paths = []string{data.Config.Path}
	}

	for _, p := range paths {
		c.data.Profiles = append(c.data.Profiles, ConfigProfile{Path: p})
	}
}
