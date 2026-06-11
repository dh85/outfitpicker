package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type testConfig struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type mockDataManager struct {
	readFunc  func(path string) ([]byte, error)
	writeFunc func(path string, data []byte) error
}

func (m *mockDataManager) Read(path string) ([]byte, error) {
	return m.readFunc(path)
}

func (m *mockDataManager) Write(path string, data []byte) error {
	return m.writeFunc(path, data)
}

type mockDirectoryProvider struct {
	baseDirFunc func() (string, error)
}

func (m *mockDirectoryProvider) BaseDirectory() (string, error) {
	return m.baseDirFunc()
}

type mockFileManager struct {
	existsFunc func(path string) bool
	removeFunc func(path string) error
	mkdirFunc  func(path string) error
}

func (m *mockFileManager) Exists(path string) bool {
	return m.existsFunc(path)
}

func (m *mockFileManager) Remove(path string) error {
	return m.removeFunc(path)
}

func (m *mockFileManager) MkdirAll(path string) error {
	return m.mkdirFunc(path)
}

func newMockDirProvider(dir string, err error) *mockDirectoryProvider {
	return &mockDirectoryProvider{
		baseDirFunc: func() (string, error) {
			return dir, err
		},
	}
}

func newMockDataManager(readData string, readErr, writeErr error) *mockDataManager {
	return &mockDataManager{
		readFunc: func(path string) ([]byte, error) {
			if readErr != nil {
				return nil, readErr
			}
			return []byte(readData), nil
		},
		writeFunc: func(path string, data []byte) error {
			return writeErr
		},
	}
}

func newMockFileManager(exists bool, removeErr, mkdirErr error) *mockFileManager {
	return &mockFileManager{
		existsFunc: func(path string) bool {
			return exists
		},
		removeFunc: func(path string) error {
			return removeErr
		},
		mkdirFunc: func(path string) error {
			return mkdirErr
		},
	}
}

func TestFileService_FilePath(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		baseErr error
		want    string
		wantErr bool
	}{
		{
			name:    "valid base directory",
			baseDir: "/Users/user/.config",
			want:    "/Users/user/.config/outfitpicker/test.json",
		},
		{
			name:    "directory provider error",
			baseErr: errors.New("no directory"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFileService[testConfig]("test.json",
				WithDirectoryProvider[testConfig](newMockDirProvider(tt.baseDir, tt.baseErr)))

			got, err := fs.FilePath()
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("FilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileService_Load(t *testing.T) {
	tests := []struct {
		name       string
		fileExists bool
		fileData   string
		readErr    error
		dirErr     error
		want       *testConfig
		wantErr    bool
	}{
		{
			name:       "file exists and valid",
			fileExists: true,
			fileData:   `{"name":"test","value":42}`,
			want:       &testConfig{Name: "test", Value: 42},
		},
		{
			name: "file does not exist",
		},
		{
			name:       "read error",
			fileExists: true,
			readErr:    errors.New("read failed"),
			wantErr:    true,
		},
		{
			name:       "missing file during read is treated as not found",
			fileExists: true,
			readErr: &os.PathError{
				Op:   "read",
				Path: "/tmp/outfitpicker/test.json",
				Err:  os.ErrNotExist,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:       "invalid json",
			fileExists: true,
			fileData:   `{invalid}`,
			wantErr:    true,
		},
		{
			name:    "directory provider error",
			dirErr:  errors.New("dir error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFileService[testConfig]("test.json",
				WithDirectoryProvider[testConfig](newMockDirProvider("/tmp", tt.dirErr)),
				WithDataManager[testConfig](newMockDataManager(tt.fileData, tt.readErr, nil)),
				WithFileManager[testConfig](newMockFileManager(tt.fileExists, nil, nil)))

			got, err := fs.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil && got != nil {
					t.Errorf("Load() = %v, want nil", got)
				}
				if tt.want != nil && (got == nil || got.Name != tt.want.Name || got.Value != tt.want.Value) {
					t.Errorf("Load() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFileService_Save(t *testing.T) {
	tests := []struct {
		name     string
		config   testConfig
		mkdirErr error
		writeErr error
		dirErr   error
		wantErr  bool
	}{
		{
			name:   "successful save",
			config: testConfig{Name: "test", Value: 42},
		},
		{
			name:     "mkdir error",
			config:   testConfig{Name: "test", Value: 42},
			mkdirErr: errors.New("mkdir failed"),
			wantErr:  true,
		},
		{
			name:     "write error",
			config:   testConfig{Name: "test", Value: 42},
			writeErr: errors.New("write failed"),
			wantErr:  true,
		},
		{
			name:    "directory provider error",
			config:  testConfig{Name: "test", Value: 42},
			dirErr:  errors.New("dir error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFileService[testConfig]("test.json",
				WithDirectoryProvider[testConfig](newMockDirProvider("/tmp", tt.dirErr)),
				WithDataManager[testConfig](newMockDataManager("", nil, tt.writeErr)),
				WithFileManager[testConfig](newMockFileManager(false, nil, tt.mkdirErr)))

			err := fs.Save(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileService_Save_DefaultDataManagerWritesAtomically(t *testing.T) {
	baseDir := t.TempDir()
	fs := NewFileService[testConfig]("test.json",
		WithDirectoryProvider[testConfig](newMockDirProvider(baseDir, nil)))

	if err := fs.Save(testConfig{Name: "test", Value: 42}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path, err := fs.FilePath()
	if err != nil {
		t.Fatalf("FilePath() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var got testConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("saved data is not valid JSON: %v", err)
	}
	if got.Name != "test" || got.Value != 42 {
		t.Fatalf("saved config = %+v, want test/42", got)
	}
	if info, err := os.Stat(path); err != nil {
		t.Fatalf("Stat() error = %v", err)
	} else if gotMode := info.Mode().Perm(); gotMode != 0o600 {
		t.Fatalf("file mode = %v, want 0600", gotMode)
	}

	tmpEntries, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".test.json.tmp-*"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(tmpEntries) != 0 {
		t.Fatalf("temporary files left behind: %v", tmpEntries)
	}
	if _, err := os.Stat(path + ".lock"); !os.IsNotExist(err) {
		t.Fatalf("lock file was not removed, stat error = %v", err)
	}
}

func TestFileService_Save_DefaultDataManagerSerializesConcurrentWrites(t *testing.T) {
	baseDir := t.TempDir()
	fs := NewFileService[testConfig]("test.json",
		WithDirectoryProvider[testConfig](newMockDirProvider(baseDir, nil)))

	const workers = 25
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs <- fs.Save(testConfig{Name: strings.Repeat("x", 512), Value: i})
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Save() concurrent error = %v", err)
		}
	}

	loaded, err := fs.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("Load() = nil, want saved config")
	}
	if loaded.Name != strings.Repeat("x", 512) || loaded.Value < 0 || loaded.Value >= workers {
		t.Fatalf("loaded config = %+v", loaded)
	}
}

func TestFileService_Save_DefaultDataManagerRecoversStaleLock(t *testing.T) {
	baseDir := t.TempDir()
	fs := NewFileService[testConfig]("test.json",
		WithDirectoryProvider[testConfig](newMockDirProvider(baseDir, nil)))
	path, err := fs.FilePath()
	if err != nil {
		t.Fatalf("FilePath() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	lockPath := path + ".lock"
	if err := os.WriteFile(lockPath, []byte("stale"), 0o600); err != nil {
		t.Fatalf("WriteFile(lock) error = %v", err)
	}
	staleTime := time.Now().Add(-staleFileLockAfter - time.Second)
	if err := os.Chtimes(lockPath, staleTime, staleTime); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	if err := fs.Save(testConfig{Name: "test", Value: 7}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := fs.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded == nil || loaded.Value != 7 {
		t.Fatalf("Load() = %+v, want value 7", loaded)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("lock file was not removed, stat error = %v", err)
	}
}

func TestDefaultDataManager_StaleCurrentProcessLockIsNotRecoverable(t *testing.T) {
	baseDir := t.TempDir()
	path := filepath.Join(baseDir, "test.json")
	lockPath := path + ".lock"
	if err := os.WriteFile(lockPath, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0o600); err != nil {
		t.Fatalf("WriteFile(lock) error = %v", err)
	}
	staleTime := time.Now().Add(-staleFileLockAfter - time.Second)
	if err := os.Chtimes(lockPath, staleTime, staleTime); err != nil {
		t.Fatalf("Chtimes() error = %v", err)
	}

	stale, err := lockIsStale(lockPath)
	if err != nil {
		t.Fatalf("lockIsStale() error = %v", err)
	}
	if stale {
		t.Fatal("lockIsStale() = true for a live process lock, want false")
	}
}

func TestFileService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		fileExists bool
		removeErr  error
		dirErr     error
		wantErr    bool
	}{
		{
			name:       "file exists and deleted",
			fileExists: true,
		},
		{
			name: "file does not exist",
		},
		{
			name:       "remove error",
			fileExists: true,
			removeErr:  errors.New("remove failed"),
			wantErr:    true,
		},
		{
			name:    "directory provider error",
			dirErr:  errors.New("dir error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFileService[testConfig]("test.json",
				WithDirectoryProvider[testConfig](newMockDirProvider("/tmp", tt.dirErr)),
				WithFileManager[testConfig](newMockFileManager(tt.fileExists, tt.removeErr, nil)))

			err := fs.Delete()
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultDirectoryProvider(t *testing.T) {
	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		defer os.Unsetenv("XDG_CONFIG_HOME")

		provider := NewDefaultDirectoryProvider()
		dir, err := provider.BaseDirectory()

		if err != nil {
			t.Errorf("BaseDirectory() error = %v", err)
		}
		if dir != "/custom/config" {
			t.Errorf("BaseDirectory() = %v, want /custom/config", dir)
		}
	})

	t.Run("uses user config dir when XDG not set", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")

		provider := NewDefaultDirectoryProvider()
		dir, err := provider.BaseDirectory()

		if err != nil {
			t.Errorf("BaseDirectory() error = %v", err)
		}
		if dir == "" {
			t.Error("BaseDirectory() returned empty string")
		}
	})
}

func TestIntegration_FileService(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewFileService[testConfig]("test.json",
		WithDirectoryProvider[testConfig](newMockDirProvider(tmpDir, nil)))

	config := testConfig{Name: "integration", Value: 99}

	if err := fs.Save(config); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path, err := fs.FilePath()
	if err != nil {
		t.Fatalf("FilePath() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if perms := info.Mode().Perm(); perms != 0o600 {
		t.Fatalf("saved file permissions = %#o, want %#o", perms, 0o600)
	}

	loaded, err := fs.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded == nil || loaded.Name != config.Name || loaded.Value != config.Value {
		t.Errorf("Load() = %v, want %v", loaded, config)
	}

	if err := fs.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	loaded, err = fs.Load()
	if err != nil {
		t.Fatalf("Load() after delete error = %v", err)
	}
	if loaded != nil {
		t.Errorf("Load() after delete = %v, want nil", loaded)
	}

	expectedPath := filepath.Join(tmpDir, "outfitpicker", "test.json")
	if path != expectedPath {
		t.Errorf("FilePath() = %v, want %v", path, expectedPath)
	}
}

type unmarshalableType struct {
	Ch chan int
}

func TestFileService_Save_MarshalError(t *testing.T) {
	fs := NewFileService[unmarshalableType]("test.json",
		WithDirectoryProvider[unmarshalableType](newMockDirProvider("/tmp", nil)),
		WithFileManager[unmarshalableType](newMockFileManager(false, nil, nil)))

	err := fs.Save(unmarshalableType{Ch: make(chan int)})
	if err == nil {
		t.Error("Save() expected error for unmarshalable type, got nil")
	}
}

func TestFileService_Save_WriteError(t *testing.T) {
	fs := NewFileService[testConfig]("test.json",
		WithDirectoryProvider[testConfig](newMockDirProvider("/tmp", nil)),
		WithDataManager[testConfig](newMockDataManager("", nil, errors.New("write failed"))),
		WithFileManager[testConfig](newMockFileManager(false, nil, nil)))

	err := fs.Save(testConfig{Name: "test", Value: 42})
	if err == nil {
		t.Error("Save() expected write error, got nil")
	}
}
