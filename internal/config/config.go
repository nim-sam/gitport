// Package config persists the basic server settings captured by `gitport init`
// in .gitport/config.json. User accounts live in the database, not here.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DirName is the per-repository configuration directory.
	DirName = ".gitport"

	// FileName is the server configuration file within DirName.
	FileName = "config.json"

	defaultHost = "0.0.0.0"
	defaultPort = "22"
)

// Config holds the basic server settings captured during initialization.
type Config struct {
	DisplayName string `json:"display_name"`
	Host        string `json:"host"`
	Port        string `json:"port"`
}

// Dir returns the .gitport directory for a repository root.
func Dir(root string) string {
	return filepath.Join(root, DirName)
}

// Path returns the path to the configuration file for a repository root.
func Path(root string) string {
	return filepath.Join(Dir(root), FileName)
}

// Exists reports whether the repository has been initialized, i.e. a
// configuration file is present.
func Exists(root string) bool {
	_, err := os.Stat(Path(root))
	return err == nil
}

// Load reads the configuration for a repository root, applying defaults for
// any field left empty.
func Load(root string) (Config, error) {
	data, err := os.ReadFile(Path(root))
	if err != nil {
		return Config{}, err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("could not parse %s: %w", Path(root), err)
	}

	c.applyDefaults()
	return c, nil
}

// Save writes the configuration for a repository root, creating the .gitport
// directory if needed.
func Save(root string, c Config) error {
	c.applyDefaults()

	if err := os.MkdirAll(Dir(root), 0700); err != nil {
		return fmt.Errorf("could not create %s: %w", Dir(root), err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode config: %w", err)
	}

	if err := os.WriteFile(Path(root), data, 0600); err != nil {
		return fmt.Errorf("could not write %s: %w", Path(root), err)
	}
	return nil
}

// applyDefaults fills unset connection settings with their standard values.
func (c *Config) applyDefaults() {
	if c.Host == "" {
		c.Host = defaultHost
	}
	if c.Port == "" {
		c.Port = defaultPort
	}
}
