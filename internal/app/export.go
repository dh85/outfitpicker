package app

import (
	"encoding/json"
	"os"
	"time"
)

// SelectionHistory represents the selection history
type SelectionHistory struct {
	Timestamp  time.Time              `json:"timestamp"`
	Categories map[string][]string    `json:"categories"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ExportManager handles import/export of selection history
type ExportManager struct{}

func NewExportManager() *ExportManager {
	return &ExportManager{}
}

func (em *ExportManager) Export(cacheData map[string][]string, filepath string) error {
	history := SelectionHistory{
		Timestamp:  time.Now(),
		Categories: cacheData,
		Metadata: map[string]interface{}{
			"version":     "1.0",
			"exported_by": "outfitpicker",
		},
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

func (em *ExportManager) Import(filepath string) (map[string][]string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var history SelectionHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return history.Categories, nil
}

func (em *ExportManager) Merge(existing, imported map[string][]string) map[string][]string {
	result := make(map[string][]string)

	// Copy existing
	for k, v := range existing {
		result[k] = append([]string(nil), v...)
	}

	// Merge imported (avoiding duplicates)
	for category, files := range imported {
		existingFiles := toSet(result[category])
		for _, file := range files {
			if !existingFiles[file] {
				result[category] = append(result[category], file)
			}
		}
	}

	return result
}
