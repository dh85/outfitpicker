package persistence

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestCacheRepository_Load(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *CacheRepository
		wantNil bool
		wantErr bool
	}{
		{
			name: "returns nil when file does not exist",
			setup: func() *CacheRepository {
				mockFS := &mockFileService[entities.OutfitCache]{
					loadResult: nil,
					loadError:  nil,
				}
				return NewCacheRepository(mockFS)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "returns cache when file exists",
			setup: func() *CacheRepository {
				cache := &entities.OutfitCache{
					Categories: make(map[string]entities.CategoryCache),
					Version:    1,
				}
				mockFS := &mockFileService[entities.OutfitCache]{
					loadResult: cache,
					loadError:  nil,
				}
				return NewCacheRepository(mockFS)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns error when load fails",
			setup: func() *CacheRepository {
				mockFS := &mockFileService[entities.OutfitCache]{
					loadResult: nil,
					loadError:  assert.AnError,
				}
				return NewCacheRepository(mockFS)
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

func TestCacheRepository_Save(t *testing.T) {
	cache := &entities.OutfitCache{
		Categories: make(map[string]entities.CategoryCache),
		Version:    1,
	}

	tests := []struct {
		name    string
		setup   func() *CacheRepository
		wantErr bool
	}{
		{
			name: "saves successfully",
			setup: func() *CacheRepository {
				mockFS := &mockFileService[entities.OutfitCache]{
					saveError: nil,
				}
				return NewCacheRepository(mockFS)
			},
			wantErr: false,
		},
		{
			name: "returns error when save fails",
			setup: func() *CacheRepository {
				mockFS := &mockFileService[entities.OutfitCache]{
					saveError: assert.AnError,
				}
				return NewCacheRepository(mockFS)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setup()
			err := repo.Save(cache)

			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCacheRepository_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *CacheRepository
		wantErr bool
	}{
		{
			name: "deletes successfully",
			setup: func() *CacheRepository {
				mockFS := &mockFileService[entities.OutfitCache]{
					deleteError: nil,
				}
				return NewCacheRepository(mockFS)
			},
			wantErr: false,
		},
		{
			name: "returns error when delete fails",
			setup: func() *CacheRepository {
				mockFS := &mockFileService[entities.OutfitCache]{
					deleteError: assert.AnError,
				}
				return NewCacheRepository(mockFS)
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
