package cli

import (
	"errors"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestRuntimeSelectionService_ShowNextUniqueRandomOutfit_SkipsExcludedCategories(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), map[string]bool{"formal": true}, nil, nil)
	categorySvc := &stubCategoryService{
		scanCategoriesResult: []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("casual", "/root/outfits/casual"), entities.CategoryStateHasOutfits, 1),
			entities.NewCategoryInfo(entities.NewCategoryReference("formal", "/root/outfits/formal"), entities.CategoryStateHasOutfits, 1),
		},
		outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {{FileName: "casual.avatar"}},
			cliTestCategoryPath("formal"): {{FileName: "formal.avatar"}},
		},
	}
	selector := NewRuntimeSelectionService(
		categorySvc,
		&stubConfigManager{config: config},
		&stubCacheManager{cache: newOutfitCachePtr()},
		NewOutfitSession(),
		func(int) int { return 0 },
	)

	outfit, err := selector.ShowNextUniqueRandomOutfit()
	if err != nil {
		t.Fatalf("ShowNextUniqueRandomOutfit() error = %v", err)
	}
	if outfit == nil {
		t.Fatal("ShowNextUniqueRandomOutfit() returned nil outfit")
	}
	if outfit.Category.Name != "casual" {
		t.Fatalf("selected category = %q, want casual", outfit.Category.Name)
	}
}

func TestRuntimeSelectionService_ShowNextUniqueRandomOutfitFrom_ResetsShownSession(t *testing.T) {
	config, _ := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), nil, nil, nil)
	categorySvc := &stubCategoryService{
		outfitsByPath: map[string][]entities.FileEntry{
			cliTestCategoryPath("casual"): {
				{FileName: "outfit1.avatar"},
				{FileName: "outfit2.avatar"},
			},
		},
	}
	session := NewOutfitSession()
	selector := NewRuntimeSelectionService(
		categorySvc,
		&stubConfigManager{config: config},
		&stubCacheManager{cache: newOutfitCachePtr()},
		session,
		func(int) int { return 0 },
	)

	first, err := selector.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("first ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}
	second, err := selector.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("second ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}
	third, err := selector.ShowNextUniqueRandomOutfitFrom("casual")
	if err != nil {
		t.Fatalf("third ShowNextUniqueRandomOutfitFrom() error = %v", err)
	}

	if first == nil || second == nil || third == nil {
		t.Fatal("expected non-nil outfits on all calls")
	}
	if first.FileName == second.FileName {
		t.Fatalf("expected second outfit to differ from first, got %q and %q", first.FileName, second.FileName)
	}
	if third.FileName != first.FileName {
		t.Fatalf("expected third outfit to reset shown session and return %q, got %q", first.FileName, third.FileName)
	}
}

func TestRuntimeSelectionService_ShowNextUniqueRandomOutfit_PropagatesSelectionErrors(t *testing.T) {
	wantErr := errors.New("config load failed")
	selector := NewRuntimeSelectionService(
		&stubCategoryService{},
		&stubConfigManager{err: wantErr},
		&stubCacheManager{cache: newOutfitCachePtr()},
		NewOutfitSession(),
		func(int) int { return 0 },
	)

	_, err := selector.ShowNextUniqueRandomOutfit()
	if !errors.Is(err, wantErr) {
		t.Fatalf("ShowNextUniqueRandomOutfit() error = %v, want %v", err, wantErr)
	}
}
