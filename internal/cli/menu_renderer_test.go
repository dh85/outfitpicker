package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestMenuRenderer_ShowWardrobeSummary(t *testing.T) {
	renderer := MenuRenderer{}
	casual := rendererCategory("casual")
	formal := rendererCategory("formal")
	archive := rendererCategory("archive")
	wardrobe := newStubWardrobeReader()
	wardrobe.outfitStates = map[string]entities.CategoryOutfitState{
		"casual":  rendererState(casual, []string{"one.avatar", "two.avatar"}, []string{"two.avatar"}, []string{"one.avatar"}),
		"formal":  rendererState(formal, []string{"jacket.avatar", "shirt.avatar"}, []string{"jacket.avatar", "shirt.avatar"}, nil),
		"archive": rendererState(archive, []string{"old.avatar"}, nil, []string{"old.avatar"}),
	}

	output := captureStdout(t, func() {
		renderer.ShowWardrobeSummary(cliTestOutfitRoot, []entities.CategoryInfo{
			entities.NewCategoryInfo(casual, entities.CategoryStateHasOutfits, 2),
			entities.NewCategoryInfo(formal, entities.CategoryStateHasOutfits, 2),
			entities.NewCategoryInfo(archive, entities.CategoryStateUserExcluded, 1),
		}, wardrobe)
	})

	assertOutputContains(t, output, "Wardrobe:", displayWardrobePath(cliTestOutfitRoot), "Progress: 2 of 5 outfits worn", "Available for random: 3", "Excluded categories: 1")
}

func TestMenuRenderer_ShowAvailableCategories(t *testing.T) {
	renderer := MenuRenderer{}
	casual := rendererCategory("casual")
	formal := rendererCategory("formal")
	wardrobe := newStubWardrobeReader()
	wardrobe.outfitStates = map[string]entities.CategoryOutfitState{
		"casual": rendererState(casual, []string{"one.avatar", "two.avatar"}, []string{"two.avatar"}, []string{"one.avatar"}),
	}
	wardrobe.outfitStateErrors = map[string]error{"formal": os.ErrNotExist}
	output := captureStdout(t, func() {
		renderer.ShowAvailableCategories([]entities.CategoryInfo{
			entities.NewCategoryInfo(casual, entities.CategoryStateHasOutfits, 2),
			entities.NewCategoryInfo(formal, entities.CategoryStateHasOutfits, 1),
		}, wardrobe)
	})

	if !strings.Contains(output, "1 of 2 outfits worn") {
		t.Fatalf("ShowAvailableCategories() output missing worn status: %q", output)
	}
	if !strings.Contains(output, "1 outfits") {
		t.Fatalf("ShowAvailableCategories() output missing fallback count status: %q", output)
	}
}

func TestMenuRenderer_ShowUnavailableCategories(t *testing.T) {
	t.Run("returns early when nothing unavailable", func(t *testing.T) {
		renderer := MenuRenderer{}
		service := NewOutfitService(newStubWardrobeReader(), &stubConfigurationController{}, &stubCommandHandler{})

		output := captureStdout(t, func() {
			renderer.ShowUnavailableCategories([]entities.CategoryInfo{
				entities.NewCategoryInfo(rendererCategory("casual"), entities.CategoryStateHasOutfits, 1),
			}, service)
		})

		if output != "" {
			t.Fatalf("ShowUnavailableCategories() output = %q, want empty", output)
		}
	})

	t.Run("renders excluded and missing outfit categories", func(t *testing.T) {
		renderer := MenuRenderer{}
		excludedCountCategory := rendererCategory("formal")
		excludedFallbackCategory := rendererCategory("archived")
		emptyCategory := rendererCategory("empty")
		noAvatarCategory := rendererCategory("docs")
		wardrobe := newStubWardrobeReader()
		wardrobe.outfitStates = map[string]entities.CategoryOutfitState{
			"formal": rendererState(excludedCountCategory, []string{"a.avatar", "b.avatar"}, []string{"b.avatar"}, []string{"a.avatar"}),
		}
		wardrobe.outfitStateErrors = map[string]error{"archived": os.ErrPermission}
		service := NewOutfitService(wardrobe, &stubConfigurationController{}, &stubCommandHandler{})

		output := captureStdout(t, func() {
			renderer.ShowUnavailableCategories([]entities.CategoryInfo{
				entities.NewCategoryInfo(excludedCountCategory, entities.CategoryStateUserExcluded, 2),
				entities.NewCategoryInfo(excludedFallbackCategory, entities.CategoryStateUserExcluded, 1),
				entities.NewCategoryInfo(emptyCategory, entities.CategoryStateEmpty, 0),
				entities.NewCategoryInfo(noAvatarCategory, entities.CategoryStateNoAvatarFiles, 0),
			}, service)
		})

		if !strings.Contains(output, "Excluded:") || !strings.Contains(output, "archived") || !strings.Contains(output, "formal (2 outfits)") {
			t.Fatalf("ShowUnavailableCategories() excluded output missing expected text: %q", output)
		}
		if !strings.Contains(output, "No outfits found:") ||
			!strings.Contains(output, "empty (Add .avatar files to "+cliTestCategoryPath("empty")+")") ||
			!strings.Contains(output, "docs (Add .avatar files to "+cliTestCategoryPath("docs")+")") {
			t.Fatalf("ShowUnavailableCategories() no-outfits output missing expected text: %q", output)
		}
	})
}

func TestMenuRenderer_ShowWornOutfits(t *testing.T) {
	renderer := MenuRenderer{}
	output := captureStdout(t, func() {
		renderer.ShowWornOutfits(map[string][]entities.OutfitReference{
			"formal": {
				entities.NewOutfitReference("jacket.avatar", rendererCategory("formal")),
			},
			"casual": {
				entities.NewOutfitReference("one.avatar", rendererCategory("casual")),
				entities.NewOutfitReference("two.avatar", rendererCategory("casual")),
			},
		})
	})

	casualIndex := strings.Index(output, "casual")
	formalIndex := strings.Index(output, "formal")
	if casualIndex == -1 || formalIndex == -1 || casualIndex > formalIndex {
		t.Fatalf("ShowWornOutfits() output not sorted by category: %q", output)
	}
	if !strings.Contains(output, "one") || !strings.Contains(output, "two") || !strings.Contains(output, "jacket") {
		t.Fatalf("ShowWornOutfits() output missing trimmed outfit names: %q", output)
	}
}

func TestMenuRenderer_ShowUnwornOutfits(t *testing.T) {
	renderer := MenuRenderer{}
	output := captureStdout(t, func() {
		renderer.ShowUnwornOutfits(map[string][]entities.OutfitReference{
			"sport": {
				entities.NewOutfitReference("shorts.avatar", rendererCategory("sport")),
			},
			"casual": {
				entities.NewOutfitReference("z.avatar", rendererCategory("casual")),
				entities.NewOutfitReference("a.avatar", rendererCategory("casual")),
			},
		})
	})

	casualIndex := strings.Index(output, "casual")
	sportIndex := strings.Index(output, "sport")
	if casualIndex == -1 || sportIndex == -1 || casualIndex > sportIndex {
		t.Fatalf("ShowUnwornOutfits() output not sorted by category: %q", output)
	}
	if !strings.Contains(output, "shorts") || !strings.Contains(output, "a") || !strings.Contains(output, "z") {
		t.Fatalf("ShowUnwornOutfits() output missing trimmed outfit names: %q", output)
	}
}

func TestMenuRenderer_ShowManualSelectionCategories(t *testing.T) {
	renderer := MenuRenderer{}
	output := captureStdout(t, func() {
		renderer.ShowManualSelectionCategories([]entities.CategoryReference{
			rendererCategory("casual"),
			rendererCategory("formal"),
		})
	})

	if !strings.Contains(output, "casual") || !strings.Contains(output, "formal") {
		t.Fatalf("ShowManualSelectionCategories() output missing categories: %q", output)
	}
}

func TestMenuRenderer_ShowManualSelectionOutfits(t *testing.T) {
	renderer := MenuRenderer{}
	category := rendererCategory("casual")
	output := captureStdout(t, func() {
		renderer.ShowManualSelectionOutfits([]entities.OutfitReference{
			entities.NewOutfitReference("one.avatar", category),
			entities.NewOutfitReference("two.avatar", category),
		}, category.Name, map[string]bool{"two.avatar": true})
	})

	if !strings.Contains(output, "one") || !strings.Contains(output, "two") || !strings.Contains(output, "(worn)") {
		t.Fatalf("ShowManualSelectionOutfits() output missing expected outfit text: %q", output)
	}
}

func TestMenuRenderer_SanitizesTerminalControlSequences(t *testing.T) {
	renderer := MenuRenderer{}
	dangerousCategory := rendererCategory("ca\x1b[2Jtual")
	output := captureStdout(t, func() {
		renderer.ShowManualSelectionOutfits([]entities.OutfitReference{
			entities.NewOutfitReference("bad\x1b[2Jname.avatar", dangerousCategory),
		}, dangerousCategory.Name, nil)
	})

	if strings.Contains(output, "\x1b[2J") {
		t.Fatalf("ShowManualSelectionOutfits() output leaked terminal clear sequence: %q", output)
	}
	if !strings.Contains(output, "bad?[2Jname") {
		t.Fatalf("ShowManualSelectionOutfits() output did not sanitize outfit name: %q", output)
	}
	if !strings.Contains(output, "ca?[2Jtual") {
		t.Fatalf("ShowManualSelectionOutfits() output did not sanitize category name: %q", output)
	}
}

func TestSanitizeTerminalText(t *testing.T) {
	if got := sanitizeTerminalText(""); got != "" {
		t.Fatalf("sanitizeTerminalText(\"\") = %q, want empty string", got)
	}

	got := sanitizeTerminalText("line\nname\t\x1b[31mred")
	if got != "line name ?[31mred" {
		t.Fatalf("sanitizeTerminalText() = %q, want %q", got, "line name ?[31mred")
	}
}

func TestMax(t *testing.T) {
	if got := max(4, 2); got != 4 {
		t.Fatalf("max(4, 2) = %d, want 4", got)
	}
	if got := max(2, 4); got != 4 {
		t.Fatalf("max(2, 4) = %d, want 4", got)
	}
}

func rendererCategory(name string) entities.CategoryReference {
	return entities.NewCategoryReference(name, cliTestCategoryPath(name))
}

func rendererState(category entities.CategoryReference, all, available, worn []string) entities.CategoryOutfitState {
	return entities.NewCategoryOutfitState(
		category,
		rendererOutfits(category, all),
		rendererOutfits(category, available),
		rendererOutfits(category, worn),
	)
}

func rendererOutfits(category entities.CategoryReference, names []string) []entities.OutfitReference {
	outfits := make([]entities.OutfitReference, 0, len(names))
	for _, name := range names {
		outfits = append(outfits, entities.NewOutfitReference(name, category))
	}
	return outfits
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer

	defer func() {
		os.Stdout = oldStdout
	}()

	run()

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("reader.Close() error = %v", err)
	}
	return string(output)
}
