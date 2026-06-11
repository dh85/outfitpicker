package usecases

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestGetCategoriesUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *GetCategoriesUseCase
		wantErr   bool
		wantCount int
		assert    func(t *testing.T, uc *GetCategoriesUseCase)
	}{
		{
			name: "gets categories successfully without exclusions",
			setup: func() *GetCategoriesUseCase {
				config, _ := entities.NewConfig("/test/path", nil, map[string]bool{"excluded": true}, nil, nil)
				return NewGetCategoriesUseCase(
					&mockCategoryService{scanResult: []entities.CategoryInfo{
						entities.NewCategoryInfo(
							entities.NewCategoryReference("casual", "/test/path/casual"),
							entities.CategoryStateHasOutfits,
							5,
						),
						entities.NewCategoryInfo(
							entities.NewCategoryReference("excluded", "/test/path/excluded"),
							entities.CategoryStateUserExcluded,
							3,
						),
					}},
					&mockConfigUseCase{loadResult: config},
				)
			},
			wantCount: 2,
			assert: func(t *testing.T, uc *GetCategoriesUseCase) {
				t.Helper()
				service, ok := uc.categoryService.(*mockCategoryService)
				if !ok {
					t.Fatal("expected mockCategoryService")
				}
				if service.lastScanRootPath != "/test/path" {
					t.Fatalf("lastScanRootPath = %q, want /test/path", service.lastScanRootPath)
				}
				if service.lastExcludedCategories == nil {
					t.Fatal("expected excluded categories to be forwarded")
				}
				if !service.lastExcludedCategories["excluded"] {
					t.Fatalf("expected excluded categories to contain %q", "excluded")
				}
			},
		},
		{
			name: "returns error when config load fails",
			setup: func() *GetCategoriesUseCase {
				return NewGetCategoriesUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadError: assert.AnError},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when config not found",
			setup: func() *GetCategoriesUseCase {
				return NewGetCategoriesUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadResult: nil},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when scan fails",
			setup: func() *GetCategoriesUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewGetCategoriesUseCase(
					&mockCategoryService{scanError: assert.AnError},
					&mockConfigUseCase{loadResult: config},
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := tt.setup()
			result, err := uc.Execute()
			if tt.wantErr {
				assertError(t, true, err)
				return
			}
			assertError(t, false, err)
			assertNil(t, false, result)
			if tt.wantCount > 0 && len(result) != tt.wantCount {
				t.Errorf("expected %d categories, got %d", tt.wantCount, len(result))
			}
			if tt.assert != nil {
				tt.assert(t, uc)
			}
		})
	}
}
