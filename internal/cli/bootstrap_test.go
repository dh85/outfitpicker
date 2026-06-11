package cli

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
	"github.com/dh85/outfitpicker/internal/infrastructure/system"
)

func TestLoadApplicationFromExistingConfig_ReturnsConfigurationNotFoundWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	deps := newProductionStyleRuntimeDependencies()

	_, err := LoadApplicationFromExistingConfig(deps)
	if !errors.Is(err, domainerrors.ErrConfigurationNotFound) {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v, want %v", err, domainerrors.ErrConfigurationNotFound)
	}
}

func TestLoadApplicationFromExistingConfig_PropagatesLoadError(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	deps := newProductionStyleRuntimeDependencies()
	fileService := system.NewFileService[entities.Config](configFileName)
	path, err := fileService.FilePath()
	if err != nil {
		t.Fatalf("FilePath() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err = LoadApplicationFromExistingConfig(deps)
	if err == nil {
		t.Fatal("LoadApplicationFromExistingConfig() expected error")
	}
}

func TestBootstrapPicker_MissingConfigStartsSetup(t *testing.T) {
	config := &Configuration{OutfitPath: cliTestOutfitRoot, Language: "en"}
	recorder := &messageRecorder{}

	picker, ok := BootstrapPicker(
		func() (string, error) {
			return "", domainerrors.ErrConfigurationNotFound
		},
		func(received Configuration) (string, error) {
			if !reflect.DeepEqual(received, *config) {
				t.Fatalf("createPicker() received %+v, want %+v", received, *config)
			}
			return "created", nil
		},
		func() *Configuration { return config },
		func() bool { return false },
		recorder.recordInfo,
		recorder.recordError,
		func(string) bool {
			t.Fatal("confirmRecovery() should not be called for first-time setup")
			return false
		},
	)

	if !ok {
		t.Fatal("BootstrapPicker() expected success")
	}
	if picker != "created" {
		t.Fatalf("BootstrapPicker() = %q, want %q", picker, "created")
	}
	if !reflect.DeepEqual(recorder.infos, []string{"First time setup"}) {
		t.Fatalf("info messages = %v, want [First time setup]", recorder.infos)
	}
	if len(recorder.errors) != 0 {
		t.Fatalf("error messages = %v, want none", recorder.errors)
	}
}

func TestBootstrapPicker_ReturnsLoadedPickerWhenExistingConfigLoads(t *testing.T) {
	recorder := &messageRecorder{}

	picker, ok := BootstrapPicker(
		func() (string, error) { return "existing", nil },
		func(Configuration) (string, error) {
			t.Fatal("createPicker() should not be called when existing picker loads")
			return "", nil
		},
		func() *Configuration {
			t.Fatal("promptForConfiguration() should not be called when existing picker loads")
			return nil
		},
		func() bool {
			t.Fatal("configFileExists() should not be called when existing picker loads")
			return false
		},
		recorder.recordInfo,
		recorder.recordError,
		func(string) bool {
			t.Fatal("confirmRecovery() should not be called when existing picker loads")
			return false
		},
	)

	if !ok {
		t.Fatal("BootstrapPicker() expected success")
	}
	if picker != "existing" {
		t.Fatalf("BootstrapPicker() = %q, want %q", picker, "existing")
	}
	if len(recorder.infos) != 0 || len(recorder.errors) != 0 {
		t.Fatalf("unexpected messages: infos=%v errors=%v", recorder.infos, recorder.errors)
	}
}

func TestBootstrapPicker_ConfigErrorWithoutFileReportsLoadError(t *testing.T) {
	recorder := &messageRecorder{}

	_, ok := BootstrapPicker(
		func() (string, error) { return "", domainerrors.ErrCache },
		func(Configuration) (string, error) {
			t.Fatal("createPicker() should not be called when config file is missing")
			return "", nil
		},
		func() *Configuration {
			t.Fatal("promptForConfiguration() should not be called when config file is missing")
			return nil
		},
		func() bool { return false },
		recorder.recordInfo,
		recorder.recordError,
		func(string) bool {
			t.Fatal("confirmRecovery() should not be called when config file is missing")
			return false
		},
	)

	if ok {
		t.Fatal("BootstrapPicker() expected failure")
	}
	if !reflect.DeepEqual(recorder.errors, []string{"Error loading config: cache error"}) {
		t.Fatalf("error messages = %v, want load error", recorder.errors)
	}
}

func TestBootstrapPicker_UnreadableExistingConfigCanBeReplaced(t *testing.T) {
	config := &Configuration{OutfitPath: cliTestOutfitRoot, Language: "en", ExcludedCategories: []string{"old"}}
	recorder := &messageRecorder{}

	picker, ok := BootstrapPicker(
		func() (string, error) {
			return "", domainerrors.ErrFileSystem
		},
		func(received Configuration) (string, error) {
			if !reflect.DeepEqual(received, *config) {
				t.Fatalf("createPicker() received %+v, want %+v", received, *config)
			}
			return "recovered", nil
		},
		func() *Configuration { return config },
		func() bool { return true },
		recorder.recordInfo,
		recorder.recordError,
		func(message string) bool {
			want := "Run setup to replace the existing config? [y/N]: "
			if message != want {
				t.Fatalf("confirmRecovery() message = %q, want %q", message, want)
			}
			return true
		},
	)

	if !ok {
		t.Fatal("BootstrapPicker() expected success")
	}
	if picker != "recovered" {
		t.Fatalf("BootstrapPicker() = %q, want %q", picker, "recovered")
	}
	if !reflect.DeepEqual(recorder.infos, []string{"Config recovery setup"}) {
		t.Fatalf("info messages = %v, want [Config recovery setup]", recorder.infos)
	}
	if !reflect.DeepEqual(recorder.errors, []string{"Existing config could not be loaded: file system error"}) {
		t.Fatalf("error messages = %v, want recovery error", recorder.errors)
	}
}

func TestBootstrapPicker_DecliningRecoveryAbortsStartup(t *testing.T) {
	recorder := &messageRecorder{}

	_, ok := BootstrapPicker(
		func() (string, error) {
			return "", domainerrors.ErrInvalidConfiguration
		},
		func(Configuration) (string, error) {
			t.Fatal("createPicker() should not be called when recovery is declined")
			return "", nil
		},
		func() *Configuration {
			t.Fatal("promptForConfiguration() should not be called when recovery is declined")
			return nil
		},
		func() bool { return true },
		recorder.recordInfo,
		recorder.recordError,
		func(string) bool { return false },
	)

	if ok {
		t.Fatal("BootstrapPicker() expected failure")
	}
	if len(recorder.infos) != 0 {
		t.Fatalf("info messages = %v, want none", recorder.infos)
	}
	if !reflect.DeepEqual(recorder.errors, []string{"Existing config could not be loaded: invalid configuration"}) {
		t.Fatalf("error messages = %v, want invalid configuration error", recorder.errors)
	}
}

func TestBootstrapPicker_UnexpectedErrorAbortsStartup(t *testing.T) {
	recorder := &messageRecorder{}

	_, ok := BootstrapPicker(
		func() (string, error) { return "", domainerrors.NewInvalidInputError("boom") },
		func(Configuration) (string, error) {
			t.Fatal("createPicker() should not be called for unexpected load errors")
			return "", nil
		},
		func() *Configuration {
			t.Fatal("promptForConfiguration() should not be called for unexpected load errors")
			return nil
		},
		func() bool {
			t.Fatal("configFileExists() should not be called for unexpected load errors")
			return false
		},
		recorder.recordInfo,
		recorder.recordError,
		func(string) bool {
			t.Fatal("confirmRecovery() should not be called for unexpected load errors")
			return false
		},
	)

	if ok {
		t.Fatal("BootstrapPicker() expected failure")
	}
	if !reflect.DeepEqual(recorder.errors, []string{"Error loading config: invalid input: boom"}) {
		t.Fatalf("error messages = %v, want boom load error", recorder.errors)
	}
}

func TestRunSetup_AbortsWhenPromptReturnsNil(t *testing.T) {
	recorder := &messageRecorder{}

	_, ok := runSetup(
		func(Configuration) (string, error) {
			t.Fatal("createPicker() should not be called when configuration prompt returns nil")
			return "", nil
		},
		func() *Configuration { return nil },
		recorder.recordError,
	)

	if ok {
		t.Fatal("runSetup() expected failure")
	}
	if len(recorder.errors) != 0 {
		t.Fatalf("error messages = %v, want none", recorder.errors)
	}
}

func TestRunSetup_ReportsCreateError(t *testing.T) {
	recorder := &messageRecorder{}
	wantErr := errors.New("create failed")

	_, ok := runSetup(
		func(Configuration) (string, error) { return "", wantErr },
		func() *Configuration { return &Configuration{OutfitPath: cliTestOutfitRoot, Language: "en"} },
		recorder.recordError,
	)

	if ok {
		t.Fatal("runSetup() expected failure")
	}
	if !reflect.DeepEqual(recorder.errors, []string{"Setup failed: create failed"}) {
		t.Fatalf("error messages = %v, want setup failure", recorder.errors)
	}
}

func TestDefaultConfigFileExists_ReturnsFalseAndTrue(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	deps := newProductionStyleRuntimeDependencies()

	if deps.ConfigExists() {
		t.Fatal("defaultConfigFileExists() = true, want false before saving config")
	}

	config, err := entities.NewConfig(cliTestOutfitRoot, stringPtr("en"), map[string]bool{"formal": true}, nil, nil)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	fileService := system.NewFileService[entities.Config](configFileName)
	if err := fileService.Save(*config); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if !deps.ConfigExists() {
		t.Fatal("defaultConfigFileExists() = false, want true after saving config")
	}
}

func TestDefaultConfigFileExists_ReturnsFalseWhenUserConfigDirUnavailable(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	deps := newProductionStyleRuntimeDependencies()

	if deps.ConfigExists() {
		t.Fatal("defaultConfigFileExists() = true, want false when config directory cannot be resolved")
	}
}

func TestCreateApplicationFromConfiguration_PersistsConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	deps := newProductionStyleRuntimeDependencies()

	app, err := CreateApplicationFromConfiguration(Configuration{
		OutfitPath:         cliTestOutfitRoot,
		Language:           "invalid",
		ExcludedCategories: []string{"formal"},
	}, deps)
	if err != nil {
		t.Fatalf("CreateApplicationFromConfiguration() error = %v", err)
	}
	appConfig, err := app.GetConfiguration()
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if appConfig.Language != "en" {
		t.Fatalf("saved language = %q, want %q", appConfig.Language, "en")
	}
	if !appConfig.ExcludedCategories["formal"] {
		t.Fatal("saved config should contain excluded category 'formal'")
	}

	loaded, err := LoadApplicationFromExistingConfig(deps)
	if err != nil {
		t.Fatalf("LoadApplicationFromExistingConfig() error = %v", err)
	}
	loadedConfig, err := loaded.GetConfiguration()
	if err != nil {
		t.Fatalf("loaded GetConfiguration() error = %v", err)
	}
	if loadedConfig.Root != cliTestOutfitRoot {
		t.Fatalf("loaded root = %q, want %q", loadedConfig.Root, cliTestOutfitRoot)
	}
	if loadedConfig.Language != "en" {
		t.Fatalf("loaded language = %q, want %q", loadedConfig.Language, "en")
	}
	if !loadedConfig.ExcludedCategories["formal"] {
		t.Fatal("loaded config should contain excluded category 'formal'")
	}
}

func TestCreateApplicationFromConfiguration_ReturnsBuildAndSaveErrors(t *testing.T) {
	t.Run("build error", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
		deps := newProductionStyleRuntimeDependencies()

		_, err := CreateApplicationFromConfiguration(Configuration{}, deps)
		if err == nil {
			t.Fatal("CreateApplicationFromConfiguration() expected build error")
		}
	})

	t.Run("save error", func(t *testing.T) {
		basePath := filepath.Join(t.TempDir(), "config-home-file")
		if err := os.WriteFile(basePath, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", basePath)

		_, err := CreateApplicationFromConfiguration(Configuration{OutfitPath: cliTestOutfitRoot, Language: "en"}, newProductionStyleRuntimeDependencies())
		if err == nil {
			t.Fatal("CreateApplicationFromConfiguration() expected save error")
		}
	})
}

func TestBootstrapApplication_FirstTimeSetupCreatesApplication(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	deps := newProductionStyleRuntimeDependencies()
	workspaceDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	outfitPath, err := os.MkdirTemp(workspaceDir, "bootstrap-test-")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(outfitPath)
	})

	withBootstrapPromptResponses(t, outfitPath, "")

	app, ok := BootstrapApplication(deps, nil)
	if !ok {
		t.Fatal("BootstrapApplication() expected success")
	}
	if app == nil {
		t.Fatalf("BootstrapApplication() returned %#v, want app with config", app)
	}
	appConfig, err := app.GetConfiguration()
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if appConfig.Root != outfitPath {
		t.Fatalf("BootstrapApplication() root = %q, want %q", appConfig.Root, outfitPath)
	}
	if !deps.ConfigExists() {
		t.Fatal("expected config file to exist after BootstrapApplication() setup")
	}
}

func TestBootstrapApplication_RecoveryFlowReplacesInvalidConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	deps := newProductionStyleRuntimeDependencies()
	fileService := system.NewFileService[entities.Config](configFileName)
	configPath, err := fileService.FilePath()
	if err != nil {
		t.Fatalf("FilePath() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(configPath, []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	workspaceDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	outfitPath, err := os.MkdirTemp(workspaceDir, "bootstrap-recovery-")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(outfitPath)
	})

	withBootstrapPromptResponses(t, "y", outfitPath, "")

	app, ok := BootstrapApplication(deps, nil)
	if !ok {
		t.Fatal("BootstrapApplication() expected recovery success")
	}
	if app == nil {
		t.Fatalf("BootstrapApplication() returned %#v, want app with config", app)
	}
	appConfig, err := app.GetConfiguration()
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if appConfig.Root != outfitPath {
		t.Fatalf("BootstrapApplication() root = %q, want %q", appConfig.Root, outfitPath)
	}
}

func withBootstrapPromptResponses(t *testing.T, responses ...string) {
	t.Helper()
	original := promptFunc
	index := 0
	promptFunc = func(string) string {
		if index >= len(responses) {
			t.Fatalf("unexpected prompt after consuming %d responses", index)
		}
		response := responses[index]
		index++
		return response
	}
	t.Cleanup(func() {
		promptFunc = original
	})
}

type messageRecorder struct {
	infos  []string
	errors []string
}

func (r *messageRecorder) recordInfo(message string) {
	r.infos = append(r.infos, message)
}

func (r *messageRecorder) recordError(message string) {
	r.errors = append(r.errors, message)
}
