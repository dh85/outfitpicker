package app

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFile handles loading/saving app configuration
type ConfigFile struct {
	path string
}

func NewConfigFile(configDir string) *ConfigFile {
	return &ConfigFile{
		path: filepath.Join(configDir, "app-config.yaml"),
	}
}

func (cf *ConfigFile) Load() (AppConfig, error) {
	config := DefaultAppConfig()

	data, err := os.ReadFile(cf.path)
	if os.IsNotExist(err) {
		return config, nil // Use defaults
	}
	if err != nil {
		return config, err
	}

	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, &config); err != nil {
		if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
			return config, err // Return YAML error as primary
		}
	}

	return config, nil
}

func (cf *ConfigFile) Save(config AppConfig) error {
	if err := os.MkdirAll(filepath.Dir(cf.path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(cf.path, data, 0644)
}
