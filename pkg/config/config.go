// Package config provides configuration management for the outfit picker application.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appName         = "outfitpicker"
	configFileName  = "config.json"
	dirPermissions  = 0o700
	filePermissions = 0o600
	jsonIndent      = "  "
)

type Config struct {
	Root string `json:"root"`
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user config dir: %w", err)
	}
	return filepath.Join(dir, appName, configFileName), nil
}

func Load() (*Config, error) {
	p, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := readConfigFile(p)
	if err != nil {
		return nil, err
	}
	return parseConfig(b)
}

func Save(c *Config) error {
	p, err := getConfigPath()
	if err != nil {
		return err
	}
	if err = ensureConfigDir(p); err != nil {
		return err
	}
	data, err := encodeConfig(c)
	if err != nil {
		return err
	}
	return writeConfigFile(p, data)
}

func Delete() error {
	p, err := getConfigPath()
	if err != nil {
		return err
	}
	return removeConfigFile(p)
}

// Helper functions
func getConfigPath() (string, error) {
	return Path()
}

func readConfigFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return b, nil
}

func parseConfig(data []byte) (*Config, error) {
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &c, nil
}

func ensureConfigDir(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	return nil
}

func encodeConfig(c *Config) ([]byte, error) {
	data, err := json.MarshalIndent(c, "", jsonIndent)
	if err != nil {
		return nil, fmt.Errorf("failed to encode config: %w", err)
	}
	return data, nil
}

func writeConfigFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func removeConfigFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	return nil
}
