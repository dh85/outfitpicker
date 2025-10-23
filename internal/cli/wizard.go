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
	"github.com/dh85/outfitpicker/internal/ui"
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
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: false}
	uiInstance := ui.NewUI(stdout, theme)

	uiInstance.Header("First Time Setup")
	uiInstance.Info("Welcome to outfitpicker! Let's set up your outfit directory.")
	fmt.Fprintln(stdout, "Please enter the full path to your Outfits root directory")

	if runtime.GOOS == "windows" {
		fmt.Fprintf(stdout, "Example: %s\n", windowsExample)
	} else {
		fmt.Fprintf(stdout, "Example: %s\n", unixExample)
	}
}

func getPathInput(br *bufio.Reader, stdout io.Writer) (string, error) {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
	uiInstance := ui.NewUI(stdout, theme)

	fmt.Fprint(stdout, "📁 Root path: ")
	line, err := readLine(br)
	if err != nil {
		return "", fmt.Errorf("no input provided; run again with --root /path/to/Outfits")
	}
	path := strings.TrimSpace(line)
	if path == "" {
		uiInstance.Warning("Please enter a non-empty path")
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
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
	uiInstance := ui.NewUI(stdout, theme)

	if !info.IsDir() {
		uiInstance.Error("Path exists but is not a directory")
		return "", true, nil
	}
	uiInstance.Success("Found existing directory")
	return finalizeSetup(path, stdout)
}

func handleNonExistentPath(path string, br *bufio.Reader, stdout io.Writer) (string, bool, error) {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
	uiInstance := ui.NewUI(stdout, theme)

	uiInstance.Warning(fmt.Sprintf("Path does not exist: %s", path))
	fmt.Fprint(stdout, "📝 Create it now? [y/N]: ")
	yn, _ := readLine(br)
	if !isYesResponse(yn) {
		return "", true, nil
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		uiInstance.Error(fmt.Sprintf("Failed to create directory: %v", err))
		return "", true, nil
	}
	uiInstance.Success("Directory created successfully")
	return finalizeSetup(path, stdout)
}

func isYesResponse(response string) bool {
	response = strings.TrimSpace(response)
	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}

func finalizeSetup(path string, stdout io.Writer) (string, bool, error) {
	theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
	uiInstance := ui.NewUI(stdout, theme)

	if err := config.Save(&config.Config{Root: path}); err != nil {
		return "", false, fmt.Errorf("failed to save config: %w", err)
	}
	if err := EnsureCacheAtRoot(path, stdout); err != nil {
		return "", false, err
	}
	uiInstance.Success("Setup completed successfully!")
	uiInstance.Info(fmt.Sprintf("Your outfit directory is set to: %s", path))
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
		theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
		uiInstance := ui.NewUI(out, theme)
		uiInstance.Info(fmt.Sprintf("Created cache at %s", cm.Path()))
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

// shouldUseColors determines if colors should be used based on environment
func shouldUseColors() bool {
	// Check if output is a terminal and colors are supported
	if term := os.Getenv("TERM"); term == "dumb" || term == "" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}
