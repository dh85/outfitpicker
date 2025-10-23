package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Plugin defines the plugin interface
type Plugin interface {
	Name() string
	SupportedExtensions() []string
	ProcessFile(path string) (FileEntry, error)
	Validate(entry FileEntry) error
}

// PluginManager manages file type plugins
type PluginManager struct {
	plugins map[string]Plugin
}

func NewPluginManager() *PluginManager {
	pm := &PluginManager{
		plugins: make(map[string]Plugin),
	}
	
	// Register built-in plugins
	pm.Register(ImagePlugin{})
	pm.Register(DocumentPlugin{})
	
	return pm
}

func (pm *PluginManager) Register(plugin Plugin) {
	pm.plugins[plugin.Name()] = plugin
}

func (pm *PluginManager) GetPlugin(filename string) Plugin {
	ext := strings.ToLower(filepath.Ext(filename))
	
	for _, plugin := range pm.plugins {
		for _, supportedExt := range plugin.SupportedExtensions() {
			if ext == supportedExt {
				return plugin
			}
		}
	}
	
	return DefaultPlugin{} // Fallback
}

// Built-in plugins
type ImagePlugin struct{}

func (p ImagePlugin) Name() string { return "image" }
func (p ImagePlugin) SupportedExtensions() []string {
	return []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
}
func (p ImagePlugin) ProcessFile(path string) (FileEntry, error) {
	return FileEntry{
		FilePath:     path,
		FileName:     filepath.Base(path),
		CategoryPath: filepath.Dir(path),
	}, nil
}
func (p ImagePlugin) Validate(entry FileEntry) error {
	if entry.FileName == "" {
		return fmt.Errorf("image file name cannot be empty")
	}
	return nil
}

type DocumentPlugin struct{}

func (p DocumentPlugin) Name() string { return "document" }
func (p DocumentPlugin) SupportedExtensions() []string {
	return []string{".pdf", ".doc", ".docx", ".txt", ".md"}
}
func (p DocumentPlugin) ProcessFile(path string) (FileEntry, error) {
	return FileEntry{
		FilePath:     path,
		FileName:     filepath.Base(path),
		CategoryPath: filepath.Dir(path),
	}, nil
}
func (p DocumentPlugin) Validate(entry FileEntry) error {
	if entry.FileName == "" {
		return fmt.Errorf("document file name cannot be empty")
	}
	return nil
}

type DefaultPlugin struct{}

func (p DefaultPlugin) Name() string { return "default" }
func (p DefaultPlugin) SupportedExtensions() []string { return []string{} }
func (p DefaultPlugin) ProcessFile(path string) (FileEntry, error) {
	return FileEntry{
		FilePath:     path,
		FileName:     filepath.Base(path),
		CategoryPath: filepath.Dir(path),
	}, nil
}
func (p DefaultPlugin) Validate(entry FileEntry) error { return nil }