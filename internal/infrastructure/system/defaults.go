package system

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var fileLocks sync.Map

const (
	fileLockPollInterval = 10 * time.Millisecond
	fileLockTimeout      = 5 * time.Second
	staleFileLockAfter   = 30 * time.Second
)

type defaultDataManager struct{}

func (d *defaultDataManager) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (d *defaultDataManager) Write(path string, data []byte) error {
	lock := lockForPath(path)
	lock.Lock()
	defer lock.Unlock()

	release, err := acquireFileLock(path)
	if err != nil {
		return err
	}
	defer release()

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	committed = true

	return syncDirectory(dir)
}

func lockForPath(path string) *sync.Mutex {
	absolute, err := filepath.Abs(path)
	if err != nil {
		absolute = filepath.Clean(path)
	}
	value, _ := fileLocks.LoadOrStore(absolute, &sync.Mutex{})
	return value.(*sync.Mutex)
}

func acquireFileLock(path string) (func(), error) {
	lockPath := path + ".lock"
	deadline := time.Now().Add(fileLockTimeout)

	for {
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			_, _ = fmt.Fprintf(lockFile, "%d\n", os.Getpid())
			return func() {
				_ = lockFile.Close()
				_ = os.Remove(lockPath)
			}, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		stale, staleErr := lockIsStale(lockPath)
		if staleErr != nil && !os.IsNotExist(staleErr) {
			return nil, staleErr
		}
		if stale {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for file lock %q", lockPath)
		}
		time.Sleep(fileLockPollInterval)
	}
}

func lockIsStale(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if time.Since(info.ModTime()) <= staleFileLockAfter {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return true, nil
	}
	return !processExists(pid), nil
}

func syncDirectory(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

type defaultDirectoryProvider struct{}

func NewDefaultDirectoryProvider() DirectoryProvider {
	return &defaultDirectoryProvider{}
}

func (d *defaultDirectoryProvider) BaseDirectory() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg, nil
	}
	return os.UserConfigDir()
}

type defaultFileManager struct{}

func (d *defaultFileManager) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (d *defaultFileManager) Remove(path string) error {
	return os.Remove(path)
}

func (d *defaultFileManager) MkdirAll(path string) error {
	return os.MkdirAll(path, 0700)
}
