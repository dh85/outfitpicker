// Package storage provides caching functionality for the outfit picker application.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	cacheFileName = "OutfitSelectorCache.json"
	cacheFileMode = 0o600
)

type Map map[string][]string

type Manager struct {
	cacheFile string
}

func NewManager(rootPath string) (*Manager, error) {
	cachePath := getCachePath(rootPath)
	if cachePath == "" {
		sys, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get cache dir: %w", err)
		}
		cachePath = filepath.Join(sys, cacheFileName)
	}
	return &Manager{cacheFile: cachePath}, nil
}

func getCachePath(rootPath string) string {
	if rootPath == "" {
		return ""
	}
	if st, err := os.Stat(rootPath); err == nil && st.IsDir() {
		return filepath.Join(rootPath, cacheFileName)
	}
	return ""
}

func (m *Manager) Path() string { return m.cacheFile }

func (m *Manager) Load() Map {
	data, err := os.ReadFile(m.cacheFile)
	if err != nil {
		return Map{}
	}
	var c Map
	if err := json.Unmarshal(data, &c); err != nil {
		return Map{}
	}
	return c
}

func (m *Manager) Save(c Map) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("error: could not encode cache:", err)
		return
	}
	if err := os.WriteFile(m.cacheFile, data, cacheFileMode); err != nil {
		fmt.Println("error: could not write cache:", err)
		return
	}
	_ = os.Chmod(m.cacheFile, cacheFileMode)
}

func (m *Manager) Add(fileName, categoryPath string) {
	c := m.Load()
	if contains(c[categoryPath], fileName) {
		return
	}
	c[categoryPath] = append(c[categoryPath], fileName)
	m.Save(c)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (m *Manager) Clear(categoryPath string) {
	c := m.Load()
	delete(c, categoryPath)
	m.Save(c)
}
