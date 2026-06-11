package usecases

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestCacheUseCase_LoadOrCreate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *CacheUseCase
		wantNil bool
		wantErr bool
	}{
		{
			name: "loads existing cache",
			setup: func() *CacheUseCase {
				cache := entities.NewOutfitCache()
				return NewCacheUseCase(&mockCacheRepo{loadResult: &cache})
			},
		},
		{
			name: "creates new cache when none exists",
			setup: func() *CacheUseCase {
				return NewCacheUseCase(&mockCacheRepo{loadResult: nil})
			},
		},
		{
			name: "returns error on load failure",
			setup: func() *CacheUseCase {
				return NewCacheUseCase(&mockCacheRepo{loadError: assert.AnError})
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setup().LoadOrCreate()
			assertError(t, tt.wantErr, err)
			if !tt.wantErr {
				assertNil(t, tt.wantNil, result)
			}
		})
	}
}

func TestCacheUseCase_Save(t *testing.T) {
	cache := entities.NewOutfitCache()

	tests := []struct {
		name    string
		setup   func() *CacheUseCase
		wantErr bool
	}{
		{
			name:  "saves successfully",
			setup: func() *CacheUseCase { return NewCacheUseCase(&mockCacheRepo{}) },
		},
		{
			name:    "returns error on save failure",
			setup:   func() *CacheUseCase { return NewCacheUseCase(&mockCacheRepo{saveError: assert.AnError}) },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().Save(&cache)
			assertError(t, tt.wantErr, err)
		})
	}
}

func TestCacheUseCase_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *CacheUseCase
		wantErr bool
	}{
		{
			name:  "deletes successfully",
			setup: func() *CacheUseCase { return NewCacheUseCase(&mockCacheRepo{}) },
		},
		{
			name:    "returns error on delete failure",
			setup:   func() *CacheUseCase { return NewCacheUseCase(&mockCacheRepo{deleteError: assert.AnError}) },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().Delete()
			assertError(t, tt.wantErr, err)
		})
	}
}
