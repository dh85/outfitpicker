package app

import (
	"testing"
)

func TestI18n_English(t *testing.T) {
	i18n := NewI18n("en")
	
	if i18n.T("category") != "Category" {
		t.Error("English translation failed")
	}
	
	// Test with arguments
	result := i18n.T("total_files", "Beach", 5)
	expected := "Total files in \"Beach\": 5"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestI18n_Spanish(t *testing.T) {
	i18n := NewI18n("es")
	
	if i18n.T("category") != "Categoría" {
		t.Error("Spanish translation failed")
	}
	
	// Test with arguments
	result := i18n.T("total_files", "Playa", 3)
	expected := "Total de archivos en \"Playa\": 3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestI18n_FallbackToEnglish(t *testing.T) {
	i18n := NewI18n("fr") // Unsupported locale
	
	// Should fallback to English
	if i18n.T("category") != "Category" {
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
	result := i18n.T("options")
	if result != "Options:" {
		t.Errorf("expected 'Options:', got %q", result)
	}
}