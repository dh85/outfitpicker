package persistence

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestConfigRepository_Load(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *ConfigRepository
		wantNil  bool
		wantErr  bool
	}{
		{
			name: "returns nil when file does not exist",
			setup: func() *ConfigRepository {
				mockFS := &mockFileService[entities.Config]{
					loadResult: nil,
					loadError:  nil,
				}
				return NewConfigRepository(mockFS)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "returns config when file exists",
			setup: func() *ConfigRepository {
				config := &entities.Config{
					Root:     "/test/path",
					Language: "en",
				}
				mockFS := &mockFileService[entities.Config]{
					loadResult: config,
					loadError:  nil,
				}
				return NewConfigRepository(mockFS)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns error when load fails",
			setup: func() *ConfigRepository {
				mockFS := &mockFileService[entities.Config]{
					loadResult: nil,
					loadError:  assert.AnError,
				}
				return NewConfigRepository(mockFS)
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setup()
			result, err := repo.Load()

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.wantNil && result != nil {
				t.Error("expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestConfigRepository_Save(t *testing.T) {
	config := &entities.Config{
		Root:     "/test/path",
		Language: "en",
	}

	tests := []struct {
		name    string
		setup   func() *ConfigRepository
		wantErr bool
	}{
		{
			name: "saves successfully",
			setup: func() *ConfigRepository {
				mockFS := &mockFileService[entities.Config]{
					saveError: nil,
				}
				return NewConfigRepository(mockFS)
			},
			wantErr: false,
		},
		{
			name: "returns error when save fails",
			setup: func() *ConfigRepository {
				mockFS := &mockFileService[entities.Config]{
					saveError: assert.AnError,
				}
				return NewConfigRepository(mockFS)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setup()
			err := repo.Save(config)

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfigRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ConfigRepository
		wantErr bool
	}{
		{
			name: "deletes successfully",
			setup: func() *ConfigRepository {
				mockFS := &mockFileService[entities.Config]{
					deleteError: nil,
				}
				return NewConfigRepository(mockFS)
			},
			wantErr: false,
		},
		{
			name: "returns error when delete fails",
			setup: func() *ConfigRepository {
				mockFS := &mockFileService[entities.Config]{
					deleteError: assert.AnError,
				}
				return NewConfigRepository(mockFS)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setup()
			err := repo.Delete()

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Mock implementations
type mockFileService[T any] struct {
	loadResult  *T
	loadError   error
	saveError   error
	deleteError error
}

func (m *mockFileService[T]) Load() (*T, error) {
	return m.loadResult, m.loadError
}

func (m *mockFileService[T]) Save(obj T) error {
	return m.saveError
}

func (m *mockFileService[T]) Delete() error {
	return m.deleteError
}

// Test helper for errors
var assert = struct {
	AnError error
}{
	AnError: &testError{},
}

type testError struct{}

func (e *testError) Error() string {
	return "test error"
}