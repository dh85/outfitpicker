package mocks

import (
	"errors"
	"strings"
)

type MockCache struct {
	Data   map[string][]string
	SaveErr error
}

func (m *MockCache) Load() map[string][]string {
	if m.Data == nil {
		return make(map[string][]string)
	}
	return m.Data
}

func (m *MockCache) Save(data map[string][]string) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Data = data
	return nil
}

func (m *MockCache) Add(filename, categoryPath string) {
	if m.Data == nil {
		m.Data = make(map[string][]string)
	}
	m.Data[categoryPath] = append(m.Data[categoryPath], filename)
}

func (m *MockCache) Clear(categoryPath string) {
	if m.Data != nil {
		delete(m.Data, categoryPath)
	}
}

type MockPrompter struct {
	Responses []string
	Index     int
	ReadErr   error
}

func (m *MockPrompter) ReadLine() (string, error) {
	if m.ReadErr != nil {
		return "", m.ReadErr
	}
	if m.Index >= len(m.Responses) {
		return "", errors.New("no more responses")
	}
	response := m.Responses[m.Index]
	m.Index++
	return response, nil
}

func (m *MockPrompter) ReadLineLower() (string, error) {
	line, err := m.ReadLine()
	return strings.ToLower(line), err
}

func (m *MockPrompter) ReadLineLowerDefault(defaultValue string) (string, error) {
	line, err := m.ReadLineLower()
	if err != nil || strings.TrimSpace(line) == "" {
		return defaultValue, nil
	}
	return line, nil
}