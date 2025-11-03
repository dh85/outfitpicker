package app

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	m := NewMetrics()

	// Test initial state
	if m.OutfitsSelected != 0 {
		t.Errorf("Expected OutfitsSelected to be 0, got %d", m.OutfitsSelected)
	}
	if m.OutfitsSkipped != 0 {
		t.Errorf("Expected OutfitsSkipped to be 0, got %d", m.OutfitsSkipped)
	}

	// Test recording
	m.RecordSelection()
	m.RecordSelection()
	m.RecordSkip()
	m.RecordCategoryVisit()

	if m.OutfitsSelected != 2 {
		t.Errorf("Expected OutfitsSelected to be 2, got %d", m.OutfitsSelected)
	}
	if m.OutfitsSkipped != 1 {
		t.Errorf("Expected OutfitsSkipped to be 1, got %d", m.OutfitsSkipped)
	}
	if m.CategoriesVisited != 1 {
		t.Errorf("Expected CategoriesVisited to be 1, got %d", m.CategoriesVisited)
	}

	// Test session duration
	time.Sleep(10 * time.Millisecond)
	duration := m.SessionDuration()
	if duration < 10*time.Millisecond {
		t.Errorf("Expected session duration >= 10ms, got %v", duration)
	}
}
