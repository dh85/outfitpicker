package app

// AppConfig holds application behavior configuration
type AppConfig struct {
	// File filtering
	ExcludedDirs     []string
	HiddenFilePrefix string
	
	// Display options
	ShowEmojis       bool
	ShowProgress     bool
	
	// Behavior
	DefaultAction    string
	AutoClearCache   bool
}

// DefaultAppConfig returns sensible defaults
func DefaultAppConfig() AppConfig {
	return AppConfig{
		ExcludedDirs:     []string{"Downloads"},
		HiddenFilePrefix: ".",
		ShowEmojis:       true,
		ShowProgress:     true,
		DefaultAction:    ActionKeep,
		AutoClearCache:   true,
	}
}