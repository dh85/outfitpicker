package cli

import (
	stderrors "errors"
	"fmt"

	domainerrors "github.com/dh85/outfitpicker/internal/domain/errors"
)

const configFileName = "config.json"
const cacheFileName = "cache.json"

func LoadApplicationFromExistingConfig(deps RuntimeDependencies) (*Application, error) {
	config, err := deps.ConfigManager.LoadOrCreate()
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domainerrors.ErrConfigurationNotFound
	}

	return buildApplication(config, deps), nil
}

func CreateApplicationFromConfiguration(configuration Configuration, deps RuntimeDependencies) (*Application, error) {
	config, err := configuration.BuildConfig()
	if err != nil {
		return nil, err
	}

	if err := deps.ConfigManager.Save(config); err != nil {
		return nil, err
	}

	return buildApplication(config, deps), nil
}

func BootstrapApplication(deps RuntimeDependencies, console Console) (*Application, bool) {
	return BootstrapPicker(
		func() (*Application, error) { return LoadApplicationFromExistingConfig(deps) },
		func(configuration Configuration) (*Application, error) {
			return CreateApplicationFromConfiguration(configuration, deps)
		},
		func() *Configuration { return PromptConfigurationWithConsole(console, deps.CategorySvc) },
		deps.ConfigExists,
		consoleOrDefault(console).Info,
		consoleOrDefault(console).Error,
		func(message string) bool { return ConfirmWithConsole(console, message, false) },
	)
}

func BootstrapPicker[T any](
	loadExistingPicker func() (T, error),
	createPicker func(Configuration) (T, error),
	promptForConfiguration func() *Configuration,
	configFileExists func() bool,
	info func(string),
	reportError func(string),
	confirmRecovery func(string) bool,
) (T, bool) {
	var zero T

	picker, err := loadExistingPicker()
	if err == nil {
		return picker, true
	}

	mappedError := domainerrors.MapError(err)
	switch {
	case stderrors.Is(mappedError, domainerrors.ErrConfigurationNotFound):
		info("First time setup")
		return runSetup(createPicker, promptForConfiguration, reportError)

	case stderrors.Is(mappedError, domainerrors.ErrInvalidConfiguration), stderrors.Is(mappedError, domainerrors.ErrCache), stderrors.Is(mappedError, domainerrors.ErrFileSystem):
		if !configFileExists() {
			reportError(fmt.Sprintf("Error loading config: %v", err))
			return zero, false
		}

		reportError(fmt.Sprintf("Existing config could not be loaded: %v", err))
		if !confirmRecovery("Run setup to replace the existing config? [y/N]: ") {
			return zero, false
		}

		info("Config recovery setup")
		return runSetup(createPicker, promptForConfiguration, reportError)

	default:
		reportError(fmt.Sprintf("Error loading config: %v", err))
		return zero, false
	}
}

func runSetup[T any](
	createPicker func(Configuration) (T, error),
	promptForConfiguration func() *Configuration,
	reportError func(string),
) (T, bool) {
	var zero T

	configuration := promptForConfiguration()
	if configuration == nil {
		return zero, false
	}

	picker, err := createPicker(*configuration)
	if err != nil {
		reportError(fmt.Sprintf("Setup failed: %v", err))
		return zero, false
	}

	return picker, true
}
