package cli

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestPrompt_UsesDefaultPromptFunc(t *testing.T) {
	t.Run("reads newline terminated input", func(t *testing.T) {
		gotOutput, gotValue := runPromptWithInput(t, "Enter path: ", cliTestOutfitRoot+"\n")
		if gotValue != cliTestOutfitRoot {
			t.Fatalf("prompt() = %q, want %q", gotValue, cliTestOutfitRoot)
		}
		if !strings.Contains(gotOutput, "Enter path: ") {
			t.Fatalf("prompt() output = %q, want prompt message", gotOutput)
		}
	})

	t.Run("returns trimmed input on eof without newline", func(t *testing.T) {
		_, gotValue := runPromptWithInput(t, "Enter path: ", cliTestOutfitRoot)
		if gotValue != cliTestOutfitRoot {
			t.Fatalf("prompt() = %q, want %q", gotValue, cliTestOutfitRoot)
		}
	})
}

func TestPromptConfiguration(t *testing.T) {
	t.Run("blank path returns nil", func(t *testing.T) {
		restore := withPromptResponses(t, "   ")
		defer restore()

		got := PromptConfiguration(nil)
		if got != nil {
			t.Fatalf("PromptConfiguration() = %#v, want nil", got)
		}
	})

	t.Run("builds configuration from prompts after confirming wardrobe", func(t *testing.T) {
		service := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("Casual", cliTestCategoryPath("Casual")), entities.CategoryStateHasOutfits, 2),
				entities.NewCategoryInfo(entities.NewCategoryReference("Formal", cliTestCategoryPath("Formal")), entities.CategoryStateEmpty, 0),
			},
		}
		restore := withPromptResponses(t, cliTestOutfitRoot, "", "fr", "2, Casual")
		defer restore()

		got := PromptConfiguration(service)
		want := &Configuration{
			OutfitPath:         cliTestOutfitRoot,
			Language:           "fr",
			ExcludedCategories: []string{"Casual", "Formal"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("PromptConfiguration() = %#v, want %#v", got, want)
		}
	})

	t.Run("returns nil when wardrobe preview is declined", func(t *testing.T) {
		service := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("Casual", cliTestCategoryPath("Casual")), entities.CategoryStateHasOutfits, 2),
			},
		}
		restore := withPromptResponses(t, cliTestOutfitRoot, "n")
		defer restore()

		if got := PromptConfiguration(service); got != nil {
			t.Fatalf("PromptConfiguration() = %#v, want nil", got)
		}
	})
}

func TestPromptConfigurationWithConsole_FirstRunOnboarding(t *testing.T) {
	service := &stubCategoryService{
		scanCategoriesResult: []entities.CategoryInfo{
			entities.NewCategoryInfo(entities.NewCategoryReference("shirts", cliTestCategoryPath("shirts")), entities.CategoryStateHasOutfits, 12),
			entities.NewCategoryInfo(entities.NewCategoryReference("shoes", cliTestCategoryPath("shoes")), entities.CategoryStateNoAvatarFiles, 0),
		},
	}
	input := strings.Join([]string{cliTestOutfitRoot, "", "", ""}, "\n") + "\n"
	var output strings.Builder

	got := PromptConfigurationWithConsole(TerminalConsole{stdin: strings.NewReader(input), stdout: &output, stderr: &output}, service)
	if got == nil {
		t.Fatal("PromptConfigurationWithConsole() = nil, want configuration")
	}
	assertOutputContains(t, output.String(),
		"Welcome to OutfitPicker",
		"No wardrobe directory is configured yet.",
		"Where are your outfits stored?",
		"Found 2 categories:",
		"shirts",
		"12 outfits",
		"shoes",
		"no .avatar files found",
		"Use this wardrobe? [Y/n]",
	)
}

func TestPromptConfigurationWithConsole_ScanErrorAbortsSetup(t *testing.T) {
	service := &stubCategoryService{scanCategoriesErr: errors.New("missing directory")}
	var output strings.Builder

	got := PromptConfigurationWithConsole(TerminalConsole{stdin: strings.NewReader(cliTestOutfitRoot + "\n"), stdout: &output, stderr: &output}, service)

	if got != nil {
		t.Fatalf("PromptConfigurationWithConsole() = %#v, want nil", got)
	}
	assertOutputContains(t, output.String(), "Could not scan wardrobe directory")
}

func runPromptWithInput(t *testing.T, message string, input string) (string, string) {
	t.Helper()

	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stdin error = %v", err)
	}
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stdout error = %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	if _, err := io.WriteString(stdinWriter, input); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := stdinWriter.Close(); err != nil {
		t.Fatalf("stdinWriter.Close() error = %v", err)
	}

	gotValue := prompt(message)

	if err := stdoutWriter.Close(); err != nil {
		t.Fatalf("stdoutWriter.Close() error = %v", err)
	}
	output, err := io.ReadAll(stdoutReader)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if err := stdoutReader.Close(); err != nil {
		t.Fatalf("stdoutReader.Close() error = %v", err)
	}
	if err := stdinReader.Close(); err != nil {
		t.Fatalf("stdinReader.Close() error = %v", err)
	}

	return string(output), gotValue
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name         string
		responses    []string
		defaultValue bool
		want         bool
	}{
		{name: "empty uses true default", responses: []string{""}, defaultValue: true, want: true},
		{name: "empty uses false default", responses: []string{""}, defaultValue: false, want: false},
		{name: "yes", responses: []string{"yes"}, defaultValue: false, want: true},
		{name: "short yes trimmed", responses: []string{" Y "}, defaultValue: false, want: true},
		{name: "no", responses: []string{"no"}, defaultValue: true, want: false},
		{name: "short no trimmed", responses: []string{" n "}, defaultValue: true, want: false},
		{name: "invalid uses default", responses: []string{"maybe"}, defaultValue: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := withPromptResponses(t, tt.responses...)
			defer restore()

			got := Confirm("Proceed? ", tt.defaultValue)
			if got != tt.want {
				t.Fatalf("Confirm() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestParseExcludedCategories(t *testing.T) {
	t.Run("blank returns nil", func(t *testing.T) {
		if got := parseExcludedCategories("   "); got != nil {
			t.Fatalf("parseExcludedCategories() = %#v, want nil", got)
		}
	})

	t.Run("trims and filters empty parts", func(t *testing.T) {
		got := parseExcludedCategories(" casual, , formal ,winter ")
		want := []string{"casual", "formal", "winter"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("parseExcludedCategories() = %#v, want %#v", got, want)
		}
	})
}

func TestPromptExcludedCategories(t *testing.T) {
	t.Run("nil service falls back to comma-separated input", func(t *testing.T) {
		restore := withPromptResponses(t, "casual, formal")
		defer restore()

		got := promptExcludedCategories(cliTestOutfitRoot, nil)
		want := []string{"casual", "formal"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("promptExcludedCategories() = %#v, want %#v", got, want)
		}
	})

	t.Run("scan error falls back to comma-separated input", func(t *testing.T) {
		service := &stubCategoryService{scanCategoriesErr: errors.New("boom")}
		restore := withPromptResponses(t, "casual")
		defer restore()

		got := promptExcludedCategories(cliTestOutfitRoot, service)
		want := []string{"casual"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("promptExcludedCategories() = %#v, want %#v", got, want)
		}
	})

	t.Run("no categories returns nil", func(t *testing.T) {
		service := &stubCategoryService{}

		got := promptExcludedCategories(cliTestOutfitRoot, service)
		if got != nil {
			t.Fatalf("promptExcludedCategories() = %#v, want nil", got)
		}
	})

	t.Run("category selection uses numbered choices", func(t *testing.T) {
		service := &stubCategoryService{
			scanCategoriesResult: []entities.CategoryInfo{
				entities.NewCategoryInfo(entities.NewCategoryReference("Casual", "/root/Casual"), entities.CategoryStateHasOutfits, 3),
				entities.NewCategoryInfo(entities.NewCategoryReference("Docs", "/root/Docs"), entities.CategoryStateNoAvatarFiles, 0),
				entities.NewCategoryInfo(entities.NewCategoryReference("Winter", "/root/Winter"), entities.CategoryStateEmpty, 0),
			},
		}
		restore := withPromptResponses(t, "2, Winter")
		defer restore()

		got := promptExcludedCategories(" "+cliTestOutfitRoot+" ", service)
		want := []string{"Docs", "Winter"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("promptExcludedCategories() = %#v, want %#v", got, want)
		}
	})
}

func TestExcludedCategoriesFromSelection(t *testing.T) {
	infos := []entities.CategoryInfo{
		entities.NewCategoryInfo(entities.NewCategoryReference("Downloads", "/root/Downloads"), entities.CategoryStateHasOutfits, 3),
		entities.NewCategoryInfo(entities.NewCategoryReference("Latex", "/root/Latex"), entities.CategoryStateNoAvatarFiles, 0),
		entities.NewCategoryInfo(entities.NewCategoryReference("Winter", "/root/Winter"), entities.CategoryStateEmpty, 0),
	}

	got := excludedCategoriesFromSelection("2, Downloads, 2", infos)
	want := []string{"Downloads", "Latex"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("excludedCategoriesFromSelection() = %v, want %v", got, want)
	}
}

func TestSetupCategorySelectionSuffix(t *testing.T) {
	tests := []struct {
		name string
		info entities.CategoryInfo
		want string
	}{
		{
			name: "has outfits",
			info: entities.NewCategoryInfo(entities.NewCategoryReference("Casual", "/root/Casual"), entities.CategoryStateHasOutfits, 4),
			want: " (4 outfits)",
		},
		{
			name: "empty",
			info: entities.NewCategoryInfo(entities.NewCategoryReference("Empty", "/root/Empty"), entities.CategoryStateEmpty, 0),
			want: " (empty)",
		},
		{
			name: "no avatars",
			info: entities.NewCategoryInfo(entities.NewCategoryReference("Docs", "/root/Docs"), entities.CategoryStateNoAvatarFiles, 0),
			want: " (no .avatar files)",
		},
		{
			name: "other state has no suffix",
			info: entities.NewCategoryInfo(entities.NewCategoryReference("Excluded", "/root/Excluded"), entities.CategoryStateUserExcluded, 0),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setupCategorySelectionSuffix(tt.info)
			if got != tt.want {
				t.Fatalf("setupCategorySelectionSuffix() = %q, want %q", got, tt.want)
			}
		})
	}
}
