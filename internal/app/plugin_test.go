package app

import (
	"testing"
)

func TestPluginManager(t *testing.T) {
	pm := NewPluginManager()
	
	tests := []struct {
		filename     string
		expectedName string
	}{
		{"image.jpg", "image"},
		{"photo.png", "image"},
		{"document.pdf", "document"},
		{"text.txt", "document"},
		{"unknown.xyz", "default"},
	}
	
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			plugin := pm.GetPlugin(tt.filename)
			if plugin.Name() != tt.expectedName {
				t.Errorf("expected plugin %s, got %s", tt.expectedName, plugin.Name())
			}
		})
	}
}

func TestImagePlugin(t *testing.T) {
	plugin := ImagePlugin{}
	
	if plugin.Name() != "image" {
		t.Error("expected name 'image'")
	}
	
	extensions := plugin.SupportedExtensions()
	if len(extensions) == 0 {
		t.Error("expected supported extensions")
	}
	
	// Test processing
	entry, err := plugin.ProcessFile("/path/to/image.jpg")
	if err != nil {
		t.Errorf("process failed: %v", err)
	}
	
	if entry.FileName != "image.jpg" {
		t.Errorf("expected filename 'image.jpg', got %s", entry.FileName)
	}
	
	// Test validation
	if err := plugin.Validate(entry); err != nil {
		t.Errorf("validation failed: %v", err)
	}
	
	// Test validation with empty filename
	emptyEntry := FileEntry{FileName: ""}
	if err := plugin.Validate(emptyEntry); err == nil {
		t.Error("expected validation error for empty filename")
	}
}

func TestCustomPlugin(t *testing.T) {
	pm := NewPluginManager()
	
	// Register custom plugin
	customPlugin := &testPlugin{}
	pm.Register(customPlugin)
	
	plugin := pm.GetPlugin("test.custom")
	if plugin.Name() != "test" {
		t.Error("custom plugin not registered correctly")
	}
}

// Test plugin implementation
type testPlugin struct{}

func (p *testPlugin) Name() string { return "test" }
func (p *testPlugin) SupportedExtensions() []string { return []string{".custom"} }
func (p *testPlugin) ProcessFile(path string) (FileEntry, error) {
	return FileEntry{FilePath: path}, nil
}
func (p *testPlugin) Validate(entry FileEntry) error { return nil }