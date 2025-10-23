package app

import (
	"testing"
)

func TestI18n_English(t *testing.T) {
	i18n := NewI18n("en")

	if i18n.T("outfit_picker") != "Outfit Picker" {
		t.Error("English translation failed")
	}

	// Test with arguments
	result := i18n.T("total_files", 5)
	expected := "Total files: 5"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestI18n_Spanish(t *testing.T) {
	i18n := NewI18n("es")

	if i18n.T("outfit_picker") != "Selector de Outfits" {
		t.Error("Spanish translation failed")
	}

	// Test with arguments
	result := i18n.T("total_files", 3)
	expected := "Total de archivos: 3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestI18n_FallbackToEnglish(t *testing.T) {
	i18n := NewI18n("xx") // Unsupported locale

	// Should fallback to English
	if i18n.T("outfit_picker") != "Outfit Picker" {
		t.Error("fallback to English failed")
	}
}

func TestI18n_MissingKey(t *testing.T) {
	i18n := NewI18n("en")

	// Should return key if translation not found
	result := i18n.T("nonexistent_key")
	if result != "nonexistent_key" {
		t.Errorf("expected key as fallback, got %q", result)
	}
}

func TestI18n_EmptyArgs(t *testing.T) {
	i18n := NewI18n("en")

	// Test key without arguments
	result := i18n.T("what_would_you_like")
	if result != "What would you like to do?" {
		t.Errorf("expected 'What would you like to do?', got %q", result)
	}
}

func TestDetectLocale(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "spanish_locale",
			envVars:  map[string]string{"LANG": "es_ES.UTF-8"},
			expected: "es",
		},
		{
			name:     "english_locale",
			envVars:  map[string]string{"LANG": "en_US.UTF-8"},
			expected: "en",
		},
		{
			name:     "fallback_to_english",
			envVars:  map[string]string{},
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			result := DetectLocale()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestI18n_German(t *testing.T) {
	i18n := NewI18n("de")

	if i18n.T("outfit_picker") != "Outfit-Auswahl" {
		t.Error("German translation failed")
	}

	// Test with arguments
	result := i18n.T("total_files", 5)
	expected := "Dateien insgesamt: 5"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestI18n_Japanese(t *testing.T) {
	i18n := NewI18n("ja")

	if i18n.T("outfit_picker") != "服装選択" {
		t.Error("Japanese translation failed")
	}

	if i18n.T("exit") != "終了" {
		t.Error("Japanese exit translation failed")
	}
}

func TestI18n_Polish(t *testing.T) {
	i18n := NewI18n("pl")

	if i18n.T("outfit_picker") != "Wybieracz Strojów" {
		t.Error("Polish translation failed")
	}

	if i18n.T("exit") != "Wyjście" {
		t.Error("Polish exit translation failed")
	}
}
