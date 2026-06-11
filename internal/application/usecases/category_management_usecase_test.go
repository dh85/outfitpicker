package usecases

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestCategoryManagementUseCase_ScanCategories(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *CategoryManagementUseCase
		wantErr bool
	}{
		{
			name: "scans categories successfully",
			setup: func() *CategoryManagementUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				categories := []entities.CategoryInfo{
					entities.NewCategoryInfo(
						entities.NewCategoryReference("casual", "/test/path/casual"),
						entities.CategoryStateHasOutfits,
						5,
					),
				}
				return NewCategoryManagementUseCase(
					&mockCategoryService{scanResult: categories},
					&mockConfigUseCase{loadResult: config},
				)
			},
		},
		{
			name: "returns error when config load fails",
			setup: func() *CategoryManagementUseCase {
				return NewCategoryManagementUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadError: assert.AnError},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when config not found",
			setup: func() *CategoryManagementUseCase {
				return NewCategoryManagementUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadResult: nil},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when scan fails",
			setup: func() *CategoryManagementUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewCategoryManagementUseCase(
					&mockCategoryService{scanError: assert.AnError},
					&mockConfigUseCase{loadResult: config},
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setup().ScanCategories()
			if tt.wantErr {
				assertError(t, true, err)
				return
			}
			assertError(t, false, err)
			assertNil(t, false, result)
		})
	}
}

func TestCategoryManagementUseCase_GetOutfits(t *testing.T) {
	tests := []struct {
		name         string
		categoryPath string
		setup        func() *CategoryManagementUseCase
		wantErr      bool
	}{
		{
			name:         "gets outfits successfully",
			categoryPath: "/test/category",
			setup: func() *CategoryManagementUseCase {
				return NewCategoryManagementUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{},
				)
			},
		},
		{
			name:         "returns error when get outfits fails",
			categoryPath: "/test/category",
			setup: func() *CategoryManagementUseCase {
				return NewCategoryManagementUseCase(
					&mockCategoryService{outfitsError: assert.AnError},
					&mockConfigUseCase{},
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setup().GetOutfits(tt.categoryPath)
			if tt.wantErr {
				assertError(t, true, err)
				return
			}
			assertError(t, false, err)
			assertNil(t, false, result)
		})
	}
}
