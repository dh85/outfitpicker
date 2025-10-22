// Package cli provides command-line interface utilities for the outfit picker.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dh85/outfitpicker/internal/storage"
	"github.com/dh85/outfitpicker/pkg/config"
)

const (
	dirPermissions = 0o700
	windowsExample = `C:\Users\you\Google\Outfits`
	unixExample    = `/Users/you/Google/Outfits`
)

func FirstRunWizard(stdin io.Reader, stdout io.Writer) (string, error) {
	br := bufio.NewReader(stdin)
	printWelcomeMessage(stdout)

	for {
		path, err := getPathInput(br, stdout)
		if err != nil {
			return "", err
		}
		if path == "" {
			continue
		}

		expandedPath, err := ExpandUserHome(path)
		if err != nil {
			fmt.Fprintf(stdout, "failed to resolve home directory: %v\n", err)
			continue
		}

		result, shouldContinue, err := handlePath(expandedPath, br, stdout)
		if err != nil {
			return "", err
		}
		if !shouldContinue {
			return result, nil
		}
	}
}

func printWelcomeMessage(stdout io.Writer) {
	fmt.Fprintln(stdout, "it looks like this is your first time running outfitpicker")
	fmt.Fprintln(stdout, "please enter the full path to your Outfits root directory")
	if runtime.GOOS == "windows" {
		fmt.Fprintln(stdout, "example:", windowsExample)
	} else {
		fmt.Fprintln(stdout, "example:", unixExample)
	}
}

func getPathInput(br *bufio.Reader, stdout io.Writer) (string, error) {
	fmt.Fprint(stdout, "root path: ")
	line, err := readLine(br)
	if err != nil {
		return "", fmt.Errorf("no input provided; run again with --root /path/to/Outfits")
	}
	path := strings.TrimSpace(line)
	if path == "" {
		fmt.Fprintln(stdout, "please enter a non-empty path")
	}
	return path, nil
}

func handlePath(path string, br *bufio.Reader, stdout io.Writer) (string, bool, error) {
	info, statErr := os.Stat(path)
	if statErr == nil {
		return handleExistingPath(path, info, stdout)
	}
	if os.IsNotExist(statErr) {
		return handleNonExistentPath(path, br, stdout)
	}
	fmt.Fprintf(stdout, "failed to access path: %v\n", statErr)
	return "", true, nil
}

func handleExistingPath(path string, info os.FileInfo, stdout io.Writer) (string, bool, error) {
	if !info.IsDir() {
		fmt.Fprintln(stdout, "path exists but is not a directory")
		return "", true, nil
	}
	return finalizeSetup(path, stdout)
}

func handleNonExistentPath(path string, br *bufio.Reader, stdout io.Writer) (string, bool, error) {
	fmt.Fprintf(stdout, "path does not exist: %s\n", path)
	fmt.Fprint(stdout, "create it now? [y/N]: ")
	yn, _ := readLine(br)
	if !isYesResponse(yn) {
		return "", true, nil
	}
	if err := os.MkdirAll(path, dirPermissions); err != nil {
		fmt.Fprintf(stdout, "failed to create directory: %v\n", err)
		return "", true, nil
	}
	return finalizeSetup(path, stdout)
}

func isYesResponse(response string) bool {
	response = strings.TrimSpace(response)
	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}

func finalizeSetup(path string, stdout io.Writer) (string, bool, error) {
	if err := config.Save(&config.Config{Root: path}); err != nil {
		return "", false, fmt.Errorf("failed to save config: %w", err)
	}
	if err := EnsureCacheAtRoot(path, stdout); err != nil {
		return "", false, err
	}
	return path, false, nil
}

func EnsureCacheAtRoot(root string, out io.Writer) error {
	cm, err := storage.NewManager(root)
	if err != nil {
		return fmt.Errorf("failed to init cache manager: %w", err)
	}
	if _, err := os.Stat(cm.Path()); os.IsNotExist(err) {
		// Ensure parent directory exists with secure permissions
		if err := os.MkdirAll(filepath.Dir(cm.Path()), dirPermissions); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}
		cm.Save(storage.Map{})
		fmt.Fprintf(out, "🗂️ created cache at %s\n", cm.Path())
	}
	return nil
}

func ExpandUserHome(p string) (string, error) {
	if p == "" {
		return p, nil
	}
	if strings.HasPrefix(p, "~") && runtime.GOOS != "windows" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(p, "~")), nil
	}
	return p, nil
}

func readLine(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF && len(s) > 0 {
			return strings.TrimSpace(s), nil
		}
		return "", err
	}
	return strings.TrimSpace(s), nil
}
