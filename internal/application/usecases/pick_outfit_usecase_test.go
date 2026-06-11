package usecases

import (
	"reflect"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestPickOutfitUseCase_LoadAvailableOutfits(t *testing.T) {
	tests := []struct {
		name         string
		categoryName string
		setup        func() *PickOutfitUseCase
		want         []entities.OutfitReference
		wantNil      bool
		wantErr      bool
	}{
		{
			name:         "returns error when category name invalid",
			categoryName: "",
			setup: func() *PickOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewPickOutfitUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantNil: true,
			wantErr: true,
		},
		{
			name:         "returns nil when no outfits found",
			categoryName: "casual",
			setup: func() *PickOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewPickOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantNil: true,
		},
		{
			name:         "returns nil when rotation is complete",
			categoryName: "casual",
			setup: func() *PickOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				categoryCache := entities.NewCategoryCache(2).Adding("outfit1.avatar").Adding("outfit2.avatar")
				cache = cache.Updating("casual", categoryCache)
				return NewPickOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{{FileName: "outfit1.avatar"}, {FileName: "outfit2.avatar"}}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantNil: true,
		},
		{
			name:         "filters worn outfits and preserves category path",
			categoryName: "casual",
			setup: func() *PickOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				categoryCache := entities.NewCategoryCache(3).Adding("outfit1.avatar")
				cache = cache.Updating("casual", categoryCache)
				return NewPickOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{{FileName: "outfit1.avatar"}, {FileName: "outfit2.avatar"}, {FileName: "outfit3.avatar"}}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			want: []entities.OutfitReference{
				entities.NewOutfitReference("outfit2.avatar", entities.NewCategoryReference("casual", "/test/path/casual")),
				entities.NewOutfitReference("outfit3.avatar", entities.NewCategoryReference("casual", "/test/path/casual")),
			},
		},
		{
			name:         "falls back to full file list when filtered pool is empty",
			categoryName: "casual",
			setup: func() *PickOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				categoryCache := entities.NewCategoryCache(3).Adding("outfit1.avatar").Adding("outfit2.avatar")
				cache = cache.Updating("casual", categoryCache)
				return NewPickOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{{FileName: "outfit1.avatar"}, {FileName: "outfit2.avatar"}, {FileName: "outfit2.avatar"}}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			want: []entities.OutfitReference{
				entities.NewOutfitReference("outfit1.avatar", entities.NewCategoryReference("casual", "/test/path/casual")),
				entities.NewOutfitReference("outfit2.avatar", entities.NewCategoryReference("casual", "/test/path/casual")),
				entities.NewOutfitReference("outfit2.avatar", entities.NewCategoryReference("casual", "/test/path/casual")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setup().LoadAvailableOutfits(tt.categoryName)
			if tt.wantErr {
				assertError(t, true, err)
				return
			}

			assertError(t, false, err)
			if tt.wantNil {
				if result != nil {
					t.Fatalf("LoadAvailableOutfits() = %#v, want nil", result)
				}
				return
			}

			if !reflect.DeepEqual(result, tt.want) {
				t.Fatalf("LoadAvailableOutfits() = %#v, want %#v", result, tt.want)
			}
		})
	}
}
