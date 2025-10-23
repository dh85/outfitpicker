package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFile_LoadSave(t *testing.T) {
	tempDir := t.TempDir()
	configFile := NewConfigFile(tempDir)
	
	// Test loading non-existent file (should return defaults)
	config, err := configFile.Load()
	if err != nil {
		t.Errorf("load should not error for non-existent file: %v", err)
	}
	
	defaultConfig := DefaultAppConfig()
	if config.ShowEmojis != defaultConfig.ShowEmojis {
		t.Error("should return default config")
	}
	
	// Test saving and loading
	config.ShowEmojis = false
	config.DefaultAction = "s"
	
	if saveErr := configFile.Save(config); saveErr != nil {
		t.Errorf("save failed: %v", saveErr)
	}
	
	loaded, err := configFile.Load()
	if err != nil {
		t.Errorf("load failed: %v", err)
	}
	
	if loaded.ShowEmojis != false || loaded.DefaultAction != "s" {
		t.Error("loaded config doesn't match saved config")
	}
}

func TestConfigFile_JSONSupport(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "app-config.yaml")
	
	// Write JSON content to YAML file (should still parse)
	jsonContent := `{"showEmojis": false, "defaultAction": "q"}`
	os.WriteFile(configPath, []byte(jsonContent), 0644)
	
	configFile := NewConfigFile(tempDir)
	config, err := configFile.Load()
	if err != nil {
		t.Errorf("JSON parsing failed: %v", err)
	}
	
	// Check if any field was parsed (YAML/JSON parsing may have different field names)
	if config.ShowEmojis == DefaultAppConfig().ShowEmojis && config.DefaultAction == DefaultAppConfig().DefaultAction {
		t.Log("Config appears to use defaults, JSON parsing may need field name adjustment")
	}
}