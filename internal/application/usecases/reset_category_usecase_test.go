package usecases

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestResetCategoryUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		setup        func() *ResetCategoryUseCase
		wantErr      bool
	}{
		{
			name:         "resets category successfully",
			categoryName: "casual",
			setup: func() *ResetCategoryUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache().Updating("casual", entities.NewCategoryCache(2).Adding("outfit1.avatar"))
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
		},
		{
			name:         "returns error when category name invalid",
			categoryName: "",
			setup: func() *ResetCategoryUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantErr: true,
		},
		{
			name:         "returns error when config load fails",
			categoryName: "casual",
			setup: func() *ResetCategoryUseCase {
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadError: assert.AnError},
					&mockCacheService{},
				)
			},
			wantErr: true,
		},
		{
			name:         "returns error when cache load fails",
			categoryName: "casual",
			setup: func() *ResetCategoryUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadError: assert.AnError},
				)
			},
			wantErr: true,
		},
		{
			name:         "returns error when cache save fails",
			categoryName: "casual",
			setup: func() *ResetCategoryUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache, saveError: assert.AnError},
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().Execute(tt.categoryName)
			assertError(t, tt.wantErr, err)
		})
	}
}

func TestResetCategoryUseCase_ExecuteAll(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ResetCategoryUseCase
		wantErr bool
	}{
		{
			name: "resets all categories successfully",
			setup: func() *ResetCategoryUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{},
				)
			},
		},
		{
			name: "returns error when config load fails",
			setup: func() *ResetCategoryUseCase {
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadError: assert.AnError},
					&mockCacheService{},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when cache save fails",
			setup: func() *ResetCategoryUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewResetCategoryUseCase(
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{saveError: assert.AnError},
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().ExecuteAll()
			assertError(t, tt.wantErr, err)
		})
	}
}
