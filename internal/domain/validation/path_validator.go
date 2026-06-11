package validation

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/dh85/outfitpicker/internal/domain/errors"
)

const maxPathLength = 4096

var restrictedPaths = []string{
	"/etc", "/usr", "/bin", "/sbin", "/System", "/private", "/var", "/tmp", "/root",
}

// ValidatePath validates a filesystem path for security issues.
func ValidatePath(path string) error {
	if err := validateCharacters(path); err != nil {
		return err
	}
	if err := validateLength(path); err != nil {
		return err
	}
	if err := validateTraversal(path); err != nil {
		return err
	}
	if err := validateRestrictedPaths(path); err != nil {
		return err
	}
	if err := validateSymlinks(path); err != nil {
		return err
	}
	return nil
}

func validateCharacters(path string) error {
	for _, c := range path {
		if unicode.IsControl(c) {
			return errors.ErrInvalidCharacters
		}
	}
	return nil
}

func validateLength(path string) error {
	if len(path) > maxPathLength {
		return errors.ErrPathTooLong
	}
	return nil
}

func validateTraversal(path string) error {
	for _, segment := range strings.FieldsFunc(path, func(r rune) bool {
		return r == filepath.Separator || r == '/' || r == '\\'
	}) {
		if segment == ".." {
			return errors.ErrPathTraversal
		}
	}

	return nil
}

func validateRestrictedPaths(path string) error {
	normalized := strings.ToLower(filepath.Clean(path))
	for _, restricted := range restrictedPaths {
		restricted = strings.ToLower(filepath.Clean(restricted))
		if normalized == restricted || strings.HasPrefix(normalized, restricted+string(filepath.Separator)) {
			return errors.ErrRestrictedPath
		}
	}
	return nil
}

func validateSymlinks(path string) error {
	cleaned := filepath.Clean(path)
	volume := filepath.VolumeName(cleaned)
	remaining := strings.TrimPrefix(cleaned, volume)
	current := ""
	if filepath.IsAbs(cleaned) {
		current = volume + string(filepath.Separator)
		remaining = strings.TrimPrefix(remaining, string(filepath.Separator))
	} else {
		current = volume
	}

	for _, part := range strings.Split(remaining, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}

		if current == "" {
			current = part
		} else if current == volume+string(filepath.Separator) {
			current = filepath.Join(current, part)
		} else {
			current = filepath.Join(current, part)
		}

		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.ErrSymlinkNotAllowed
		}
	}

	return nil
}

// MaxPathLength returns the maximum allowed path length.
func MaxPathLength() int {
	return maxPathLength
}

// RestrictedPaths returns the list of restricted paths.
func RestrictedPaths() []string {
	return restrictedPaths
}
