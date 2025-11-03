package app

import (
	"log"
	"os"
	"time"
)

// Metrics tracks user session statistics
type Metrics struct {
	OutfitsSelected   int64
	OutfitsSkipped    int64
	SessionStart      time.Time
	CategoriesVisited int
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		SessionStart: time.Now(),
	}
}

// RecordSelection increments the selected count
func (m *Metrics) RecordSelection() {
	m.OutfitsSelected++
}

// RecordSkip increments the skipped count
func (m *Metrics) RecordSkip() {
	m.OutfitsSkipped++
}

// RecordCategoryVisit increments the categories visited count
func (m *Metrics) RecordCategoryVisit() {
	m.CategoriesVisited++
}

// SessionDuration returns the current session duration
func (m *Metrics) SessionDuration() time.Duration {
	return time.Since(m.SessionStart)
}

// LogSession logs the session metrics
func (m *Metrics) LogSession() {
	// Only log in debug mode to avoid CI noise
	if os.Getenv("DEBUG") != "" {
		log.Printf("Session: selected=%d, skipped=%d, duration=%v, categories=%d",
			m.OutfitsSelected, m.OutfitsSkipped, m.SessionDuration(), m.CategoriesVisited)
	}
}
