package usecases

import (
	stderrors "errors"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

func TestWearOutfitUseCase_Execute(t *testing.T) {
	tests := []struct {
		name                  string
		outfit                entities.OutfitReference
		setup                 func() *WearOutfitUseCase
		wantErr               bool
		wantErrType           error
		wantRotationCompleted bool
	}{
		{
			name: "marks outfit as worn successfully",
			outfit: entities.NewOutfitReference(
				"outfit1.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
						{FileName: "outfit2.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
		},
		{
			name: "returns error when outfit invalid",
			outfit: entities.NewOutfitReference(
				"",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewWearOutfitUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when config load fails",
			outfit: entities.NewOutfitReference(
				"outfit1.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				return NewWearOutfitUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadError: assert.AnError},
					&mockCacheService{},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when cache load fails",
			outfit: entities.NewOutfitReference(
				"outfit1.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewWearOutfitUseCase(
					&mockCategoryService{},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadError: assert.AnError},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when outfit not found",
			outfit: entities.NewOutfitReference(
				"missing.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantErr:     true,
			wantErrType: domainerrors.ErrNoOutfitsAvailable,
		},
		{
			name: "returns nil when outfit already worn",
			outfit: entities.NewOutfitReference(
				"outfit1.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache().Updating("casual", entities.NewCategoryCache(2).Adding("outfit1.avatar"))
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
						{FileName: "outfit2.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
		},
		{
			name: "throws rotation completed error when rotation complete",
			outfit: entities.NewOutfitReference(
				"outfit2.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache().Updating("casual", entities.NewCategoryCache(2).Adding("outfit1.avatar"))
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
						{FileName: "outfit2.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantErr:               true,
			wantRotationCompleted: true,
		},
		{
			name: "returns error when get outfits fails",
			outfit: entities.NewOutfitReference(
				"outfit1.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsError: assert.AnError},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache},
				)
			},
			wantErr: true,
		},
		{
			name: "returns error when cache save fails",
			outfit: entities.NewOutfitReference(
				"outfit1.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache()
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
						{FileName: "outfit2.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache, saveError: assert.AnError},
				)
			},
			wantErr: true,
		},
		{
			name: "does not attempt a second save when rotation completes",
			outfit: entities.NewOutfitReference(
				"outfit2.avatar",
				entities.NewCategoryReference("casual", "/test/path/casual"),
			),
			setup: func() *WearOutfitUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				cache := entities.NewOutfitCache().Updating("casual", entities.NewCategoryCache(2).Adding("outfit1.avatar"))
				return NewWearOutfitUseCase(
					&mockCategoryService{outfitsResult: []entities.FileEntry{
						{FileName: "outfit1.avatar", IsDirectory: false},
						{FileName: "outfit2.avatar", IsDirectory: false},
					}},
					&mockConfigUseCase{loadResult: config},
					&mockCacheService{loadResult: &cache, saveErrors: []error{nil, assert.AnError}},
				)
			},
			wantErr:               true,
			wantRotationCompleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().Execute(tt.outfit)
			if tt.wantErr {
				assertError(t, true, err)
				if tt.wantRotationCompleted {
					var rotationCompleted *domainerrors.RotationCompletedError
					if !stderrors.As(err, &rotationCompleted) {
						t.Fatalf("expected RotationCompletedError, got %v", err)
					}
				}
				if tt.wantErrType != nil && err != tt.wantErrType {
					t.Errorf("expected error %v, got %v", tt.wantErrType, err)
				}
			} else {
				assertError(t, false, err)
			}
		})
	}
}
