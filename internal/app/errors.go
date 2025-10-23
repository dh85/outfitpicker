package app

import "fmt"

// AppError represents application-specific errors
type AppError struct {
	Type    string
	Message string
	Cause   error
}

func (e AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error constructors
func NewCategoryError(msg string, cause error) error {
	return AppError{Type: "CategoryError", Message: msg, Cause: cause}
}

func NewFileSystemError(msg string, cause error) error {
	return AppError{Type: "FileSystemError", Message: msg, Cause: cause}
}

func NewValidationError(msg string) error {
	return AppError{Type: "ValidationError", Message: msg}
}