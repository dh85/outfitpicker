package cli

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestOutfitService_GetAvailableOutfits(t *testing.T) {
	category := outfitServiceCategory("casual")
	wantOutfits := []entities.OutfitReference{
		entities.NewOutfitReference("one.avatar", category),
	}

	t.Run("success", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.availableOutfits = wantOutfits
		service := newStubOutfitService(picker)

		got, err := service.GetAvailableOutfits(category)
		if err != nil {
			t.Fatalf("GetAvailableOutfits() error = %v", err)
		}
		if !reflect.DeepEqual(got, wantOutfits) {
			t.Fatalf("GetAvailableOutfits() = %#v, want %#v", got, wantOutfits)
		}
	})

	t.Run("error", func(t *testing.T) {
		wantErr := errors.New("boom")
		picker := newStubRuntime()
		picker.wardrobe.availableOutfitsErr = wantErr
		service := newStubOutfitService(picker)

		_, err := service.GetAvailableOutfits(category)
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetAvailableOutfits() error = %v, want %v", err, wantErr)
		}
	})
}

func TestOutfitService_GetActualOutfitCount(t *testing.T) {
	category := outfitServiceCategory("casual")

	t.Run("success", func(t *testing.T) {
		picker := newStubRuntime()
		picker.wardrobe.outfitState = outfitServiceState(category, []string{"one.avatar", "two.avatar"}, []string{"two.avatar"}, []string{"one.avatar"})
		service := newStubOutfitService(picker)

		got, err := service.GetActualOutfitCount(category)
		if err != nil {
			t.Fatalf("GetActualOutfitCount() error = %v", err)
		}
		if got != 2 {
			t.Fatalf("GetActualOutfitCount() = %d, want 2", got)
		}
	})

	t.Run("error", func(t *testing.T) {
		wantErr := errors.New("boom")
		picker := newStubRuntime()
		picker.wardrobe.outfitStateErr = wantErr
		service := newStubOutfitService(picker)

		_, err := service.GetActualOutfitCount(category)
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetActualOutfitCount() error = %v, want %v", err, wantErr)
		}
	})
}

func TestOutfitService_GetWornOutfits(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		wantErr := errors.New("boom")
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStatesErr = wantErr
		service := newStubOutfitService(picker)

		_, err := service.GetWornOutfits()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetWornOutfits() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("filters empty categories and sorts worn outfits", func(t *testing.T) {
		casual := outfitServiceCategory("casual")
		formal := outfitServiceCategory("formal")
		sport := outfitServiceCategory("sport")
		states := map[string]entities.CategoryOutfitState{
			"casual": outfitServiceState(casual, []string{"a.avatar", "z.avatar"}, nil, []string{"z.avatar", "a.avatar"}),
			"formal": outfitServiceState(formal, []string{"jacket.avatar"}, []string{"jacket.avatar"}, nil),
			"sport":  outfitServiceState(sport, []string{"mid.avatar"}, nil, []string{"mid.avatar"}),
		}
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStates = states
		service := newStubOutfitService(picker)

		got, err := service.GetWornOutfits()
		if err != nil {
			t.Fatalf("GetWornOutfits() error = %v", err)
		}

		want := map[string][]entities.OutfitReference{
			"casual": {
				entities.NewOutfitReference("a.avatar", casual),
				entities.NewOutfitReference("z.avatar", casual),
			},
			"sport": {
				entities.NewOutfitReference("mid.avatar", sport),
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GetWornOutfits() = %#v, want %#v", got, want)
		}
	})
}

func TestOutfitService_GetUnwornOutfits(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		wantErr := errors.New("boom")
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStatesErr = wantErr
		service := newStubOutfitService(picker)

		_, err := service.GetUnwornOutfits()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetUnwornOutfits() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("filters empty categories and sorts available outfits", func(t *testing.T) {
		casual := outfitServiceCategory("casual")
		formal := outfitServiceCategory("formal")
		states := map[string]entities.CategoryOutfitState{
			"casual": outfitServiceState(casual, []string{"a.avatar", "z.avatar"}, []string{"z.avatar", "a.avatar"}, nil),
			"formal": outfitServiceState(formal, []string{"jacket.avatar"}, nil, []string{"jacket.avatar"}),
		}
		picker := newStubRuntime()
		picker.wardrobe.allOutfitStates = states
		service := newStubOutfitService(picker)

		got, err := service.GetUnwornOutfits()
		if err != nil {
			t.Fatalf("GetUnwornOutfits() error = %v", err)
		}

		want := map[string][]entities.OutfitReference{
			"casual": {
				entities.NewOutfitReference("a.avatar", casual),
				entities.NewOutfitReference("z.avatar", casual),
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GetUnwornOutfits() = %#v, want %#v", got, want)
		}
	})
}

func TestOutfitService_GetAvailableCategories(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		wantErr := errors.New("boom")
		picker := newStubRuntime()
		picker.wardrobe.categoryInfoErr = wantErr
		service := newStubOutfitService(picker)

		_, err := service.GetAvailableCategories()
		if !errors.Is(err, wantErr) {
			t.Fatalf("GetAvailableCategories() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("filters categories with outfits", func(t *testing.T) {
		infos := []entities.CategoryInfo{
			entities.NewCategoryInfo(outfitServiceCategory("casual"), entities.CategoryStateHasOutfits, 2),
			entities.NewCategoryInfo(outfitServiceCategory("formal"), entities.CategoryStateEmpty, 0),
			entities.NewCategoryInfo(outfitServiceCategory("excluded"), entities.CategoryStateUserExcluded, 1),
			entities.NewCategoryInfo(outfitServiceCategory("sport"), entities.CategoryStateHasOutfits, 1),
		}
		picker := newStubRuntime()
		picker.wardrobe.categoryInfos = infos
		service := newStubOutfitService(picker)

		got, err := service.GetAvailableCategories()
		if err != nil {
			t.Fatalf("GetAvailableCategories() error = %v", err)
		}

		want := []entities.CategoryInfo{infos[0], infos[3]}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GetAvailableCategories() = %#v, want %#v", got, want)
		}
	})
}

func outfitServiceCategory(name string) entities.CategoryReference {
	return entities.NewCategoryReference(name, cliTestCategoryPath(name))
}

func outfitServiceState(category entities.CategoryReference, all, available, worn []string) entities.CategoryOutfitState {
	return entities.NewCategoryOutfitState(
		category,
		outfitServiceOutfits(category, all),
		outfitServiceOutfits(category, available),
		outfitServiceOutfits(category, worn),
	)
}

func outfitServiceOutfits(category entities.CategoryReference, names []string) []entities.OutfitReference {
	outfits := make([]entities.OutfitReference, 0, len(names))
	for _, name := range names {
		outfits = append(outfits, entities.NewOutfitReference(name, category))
	}
	return outfits
}
