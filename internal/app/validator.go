package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Validator handles input validation
type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateRootPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return NewValidationError("root path cannot be empty")
	}
	
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return NewValidationError(fmt.Sprintf("root path does not exist: %s", path))
	}
	if err != nil {
		return NewFileSystemError("failed to access root path", err)
	}
	if !info.IsDir() {
		return NewValidationError(fmt.Sprintf("root path is not a directory: %s", path))
	}
	
	return nil
}

func (v *Validator) ValidateCategoryName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewValidationError("category name cannot be empty")
	}
	
	// Check for invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		if strings.Contains(name, char) {
			return NewValidationError(fmt.Sprintf("category name contains invalid character: %s", char))
		}
	}
	
	return nil
}

func (v *Validator) ValidateUserAction(action string) error {
	validActions := []string{ActionKeep, ActionSkip, ActionQuit, ActionRandom, ActionShowSelected, ActionShowUnselected}
	action = strings.ToLower(strings.TrimSpace(action))
	
	for _, valid := range validActions {
		if action == valid {
			return nil
		}
	}
	
	return NewValidationError(fmt.Sprintf("invalid action: %s (valid: %s)", action, strings.Join(validActions, ", ")))
}

func (v *Validator) ValidateFileExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewValidationError(fmt.Sprintf("file does not exist: %s", filepath.Base(path)))
	} else if err != nil {
		return NewFileSystemError("failed to access file", err)
	}
	return nil
}