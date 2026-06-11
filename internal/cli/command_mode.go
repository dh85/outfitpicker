package cli

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/dh85/outfitpicker/internal/domain/entities"
)

type CommandRuntime interface {
	WardrobeReader
	ConfigurationController
	OutfitCommandHandler
	RandomOutfitSelector
}

var commandRandomIndex = rand.Intn

// ExecuteCommand runs a non-interactive command. It returns handled=false when
// args should fall through to the interactive menu.
func ExecuteCommand(args []string, runtime CommandRuntime, console Console) (handled bool, exitCode int) {
	if len(args) == 0 {
		return false, 0
	}
	console = consoleOrDefault(console)
	cli := commandCLI{}
	parser, err := newCommandParser(&cli, console)
	if err != nil {
		console.Error(fmt.Sprintf("Failed to initialize command parser: %v", err))
		return true, 1
	}

	if len(args) == 1 && args[0] == "help" {
		return true, parseCommandHelp(parser, nil)
	}

	ctx, code, done := parseCommandContext(parser, args, console)
	if done {
		return true, code
	}
	if runtime == nil {
		return false, 0
	}

	commands := commandExecutor{
		runtime: runtime,
		service: NewOutfitService(runtime, runtime, runtime),
		console: console,
	}
	if err := ctx.Run(&commands); err != nil {
		return true, commandExitCode(err, console)
	}
	return true, 0
}

type commandCLI struct {
	Pick   pickCommand   `cmd:"" help:"Pick a random outfit and optionally mark it worn."`
	List   listCommand   `cmd:"" help:"List categories or outfit rotation state."`
	Reset  resetCommand  `cmd:"" help:"Reset worn outfit rotation state."`
	Config configCommand `cmd:"" help:"Show or update configuration."`
}

type pickCommand struct {
	Category        string `help:"Pick from a specific category." placeholder:"NAME"`
	IncludeExcluded bool   `help:"Include categories excluded from global random selection."`
	MarkWorn        bool   `help:"Mark the picked outfit worn without prompting." xor:"mark-mode"`
	NoMark          bool   `help:"Do not mark the picked outfit worn." xor:"mark-mode"`
}

func (c pickCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.pick(pickOptionsFromCommand(c)))
}

type listCommand struct {
	Categories listCategoriesCommand `cmd:"" help:"List wardrobe categories."`
	Worn       listWornCommand       `cmd:"" help:"List outfits already worn."`
	Unworn     listUnwornCommand     `cmd:"" help:"List outfits not yet worn."`
}

type listCategoriesCommand struct{}

func (c listCategoriesCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.listCategories())
}

type listWornCommand struct{}

func (c listWornCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.listOutfits(true))
}

type listUnwornCommand struct{}

func (c listUnwornCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.listOutfits(false))
}

type resetCommand struct {
	Category string `help:"Reset one category instead of all categories." placeholder:"NAME"`
}

func (c resetCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.reset(resetOptions{categoryName: c.Category}))
}

type configCommand struct {
	Get     configGetCommand     `cmd:"" help:"Show current configuration."`
	SetRoot configSetRootCommand `cmd:"" name:"set-root" help:"Set the wardrobe root directory."`
	Exclude configExcludeCommand `cmd:"" help:"Add categories to the exclusion list."`
}

type configGetCommand struct{}

func (c configGetCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.configGet())
}

type configSetRootCommand struct {
	Root string `arg:"" help:"New wardrobe root directory." placeholder:"PATH"`
}

func (c configSetRootCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.configSetRoot(c.Root))
}

type configExcludeCommand struct {
	Categories []string `arg:"" help:"Categories to exclude." placeholder:"CATEGORY"`
}

func (c configExcludeCommand) Run(executor *commandExecutor) error {
	return commandExit(executor.configExclude(c.Categories))
}

func newCommandParser(cli *commandCLI, console Console) (*kong.Kong, error) {
	return kong.New(
		cli,
		kong.Name("outfitpicker"),
		kong.Description("Terminal app for choosing outfits from a local wardrobe directory."),
		kong.Writers(commandOutput(console, false), commandOutput(console, true)),
		kong.Exit(func(int) { panic(commandParserExit{}) }),
	)
}

type commandParserExit struct{}

func parseCommandHelp(parser *kong.Kong, args []string) int {
	if args == nil {
		args = []string{"--help"}
	}
	_, code, done := parseCommandContext(parser, args, nil)
	if done {
		return code
	}
	return 0
}

func parseCommandContext(parser *kong.Kong, args []string, console Console) (ctx *kong.Context, code int, done bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if _, ok := recovered.(commandParserExit); !ok {
				panic(recovered)
			}
			ctx = nil
			code = 0
			done = true
		}
	}()

	parsed, err := parser.Parse(args)
	if err != nil {
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			_ = parseErr.Context.PrintUsage(true)
			if console != nil {
				console.Error("Usage: outfitpicker <command>")
			}
			return nil, 2, true
		}
		if console != nil {
			console.Error(err.Error())
		}
		return nil, 2, true
	}
	return parsed, 0, false
}

func commandOutput(console Console, stderr bool) io.Writer {
	if terminal, ok := console.(TerminalConsole); ok {
		if stderr {
			return terminal.errorOutput()
		}
		return terminal.output()
	}
	return consoleWriter{console: console}
}

type consoleWriter struct {
	console Console
}

func (w consoleWriter) Write(p []byte) (int, error) {
	w.console.Printf("%s", string(p))
	return len(p), nil
}

type commandExitError struct {
	code int
}

func (e commandExitError) Error() string {
	return fmt.Sprintf("command exited with code %d", e.code)
}

func commandExit(code int) error {
	if code == 0 {
		return nil
	}
	return commandExitError{code: code}
}

func commandExitCode(err error, console Console) int {
	var exitErr commandExitError
	if errors.As(err, &exitErr) {
		return exitErr.code
	}
	console.Error(err.Error())
	return 1
}

type commandExecutor struct {
	runtime CommandRuntime
	service OutfitService
	console Console
}

func (e commandExecutor) pick(options pickOptions) int {
	outfit, err := e.pickOutfit(options)
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to pick outfit: %v", err))
		return 1
	}
	if outfit == nil {
		e.console.Info("No outfits available")
		return 0
	}

	e.showPickedOutfit(*outfit)
	shouldMark, ok := e.shouldMarkPickedOutfit(options)
	if !ok {
		e.console.Error("Please answer yes or no")
		return 2
	}
	if !shouldMark {
		e.console.Info("Not marked worn")
		return 0
	}
	if err := e.service.WearOutfit(*outfit); err != nil {
		e.console.Error(fmt.Sprintf("Failed to mark outfit worn: %v", err))
		return 1
	}
	e.console.Success("Marked worn")
	return 0
}

func (e commandExecutor) pickOutfit(options pickOptions) (*entities.OutfitReference, error) {
	if options.categoryName != "" {
		return e.runtime.ShowNextUniqueRandomOutfitFrom(options.categoryName)
	}
	if options.includeExcluded {
		return e.pickIncludingExcludedCategories()
	}
	return e.runtime.ShowNextUniqueRandomOutfit()
}

func (e commandExecutor) pickIncludingExcludedCategories() (*entities.OutfitReference, error) {
	infos, err := e.service.GetCategoryInfo()
	if err != nil {
		return nil, err
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Category.Name < infos[j].Category.Name
	})
	var available []entities.OutfitReference
	for _, info := range infos {
		if info.State != entities.CategoryStateHasOutfits && info.State != entities.CategoryStateUserExcluded {
			continue
		}
		outfits, err := e.service.GetAvailableOutfits(info.Category)
		if err != nil {
			return nil, err
		}
		available = append(available, outfits...)
	}
	if len(available) == 0 {
		return nil, nil
	}
	return &available[commandRandomIndex(len(available))], nil
}

func (e commandExecutor) showPickedOutfit(outfit entities.OutfitReference) {
	e.console.Println("👗 Outfit picked")
	e.console.Println()
	e.console.Printf("Category: %s\n", sanitizeTerminalText(outfit.Category.Name))
	e.console.Printf("Outfit:   %s\n", sanitizeTerminalText(outfit.FileName))
	e.console.Printf("Path:     %s\n", sanitizeTerminalText(outfit.FilePath()))
	e.console.Println()
}

func (e commandExecutor) shouldMarkPickedOutfit(options pickOptions) (bool, bool) {
	switch options.markMode {
	case pickMarkNever:
		return false, true
	case pickMarkAlways:
		return true, true
	default:
		input := strings.ToLower(strings.TrimSpace(e.console.Prompt("Mark as worn? [Y/n]: ")))
		switch input {
		case "", "y", "yes":
			return true, true
		case "n", "no":
			return false, true
		default:
			return false, false
		}
	}
}

func (e commandExecutor) listCategories() int {
	infos, err := e.service.GetCategoryInfo()
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to list categories: %v", err))
		return 1
	}
	if len(infos) == 0 {
		e.console.Info("No categories found")
		return 0
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Category.Name < infos[j].Category.Name
	})
	for _, info := range infos {
		outfitWord := "outfits"
		if info.OutfitCount == 1 {
			outfitWord = "outfit"
		}
		e.console.Printf("%s\t%s\t%d %s\n", sanitizeTerminalText(info.Category.Name), info.State, info.OutfitCount, outfitWord)
	}
	return 0
}

func (e commandExecutor) listOutfits(worn bool) int {
	var outfits map[string][]entities.OutfitReference
	var err error
	if worn {
		outfits, err = e.service.GetWornOutfits()
	} else {
		outfits, err = e.service.GetUnwornOutfits()
	}
	if err != nil {
		label := "worn"
		if !worn {
			label = "unworn"
		}
		e.console.Error(fmt.Sprintf("Failed to list %s outfits: %v", label, err))
		return 1
	}
	if len(outfits) == 0 {
		if worn {
			e.console.Info("No worn outfits found")
		} else {
			e.console.Info("No unworn outfits found")
		}
		return 0
	}
	for _, category := range sortedCategoryNames(outfits) {
		e.console.Printf("%s\n", sanitizeTerminalText(category))
		for _, outfit := range outfits[category] {
			e.console.Printf("  %s\n", sanitizeTerminalText(outfit.FileName))
		}
	}
	return 0
}

type resetOptions struct {
	categoryName string
}

func (e commandExecutor) reset(options resetOptions) int {
	if options.categoryName == "" {
		if err := e.service.ResetAllCategories(); err != nil {
			e.console.Error(fmt.Sprintf("Failed to reset worn outfits: %v", err))
			return 1
		}
		e.console.Success("Reset all worn outfits")
		return 0
	}
	if err := e.service.ResetCategory(options.categoryName); err != nil {
		e.console.Error(fmt.Sprintf("Failed to reset category: %v", err))
		return 1
	}
	e.console.Success(fmt.Sprintf("Reset worn outfits for %s", options.categoryName))
	return 0
}

func (e commandExecutor) configGet() int {
	config, err := e.service.GetConfiguration()
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to load configuration: %v", err))
		return 1
	}
	e.console.Printf("Root: %s\n", sanitizeTerminalText(config.Root))
	e.console.Printf("Language: %s\n", sanitizeTerminalText(config.Language))
	excluded := sortedEnabledKeys(config.ExcludedCategories)
	if len(excluded) == 0 {
		e.console.Println("Excluded: none")
	} else {
		e.console.Printf("Excluded: %s\n", sanitizeTerminalText(strings.Join(excluded, ", ")))
	}
	return 0
}

func (e commandExecutor) configSetRoot(root string) int {
	config, err := e.service.GetConfiguration()
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to load configuration: %v", err))
		return 1
	}
	expandedRoot, err := expandHomePath(root)
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to expand path: %v", err))
		return 1
	}
	updated, err := buildUpdatedConfig(config, expandedRoot, config.Language, cloneExcludedCategories(config.ExcludedCategories))
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to update path: %v", err))
		return 1
	}
	if err := e.service.UpdateConfiguration(updated); err != nil {
		e.console.Error(fmt.Sprintf("Failed to update path: %v", err))
		return 1
	}
	e.console.Success(fmt.Sprintf("Outfit path updated to: %s", expandedRoot))
	return 0
}

func (e commandExecutor) configExclude(categories []string) int {
	config, err := e.service.GetConfiguration()
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to load configuration: %v", err))
		return 1
	}
	excluded := cloneExcludedCategories(config.ExcludedCategories)
	for _, category := range categories {
		name := strings.TrimSpace(category)
		if name != "" {
			excluded[name] = true
		}
	}
	updated, err := buildUpdatedConfig(config, config.Root, config.Language, excluded)
	if err != nil {
		e.console.Error(fmt.Sprintf("Failed to update excluded categories: %v", err))
		return 1
	}
	if err := e.service.UpdateConfiguration(updated); err != nil {
		e.console.Error(fmt.Sprintf("Failed to update excluded categories: %v", err))
		return 1
	}
	e.console.Success(fmt.Sprintf("Excluded categories updated: %s", strings.Join(sortedEnabledKeys(excluded), ", ")))
	return 0
}

type pickMarkMode int

const (
	pickMarkPrompt pickMarkMode = iota
	pickMarkAlways
	pickMarkNever
)

type pickOptions struct {
	categoryName    string
	includeExcluded bool
	markMode        pickMarkMode
}

func pickOptionsFromCommand(command pickCommand) pickOptions {
	markMode := pickMarkPrompt
	if command.MarkWorn {
		markMode = pickMarkAlways
	}
	if command.NoMark {
		markMode = pickMarkNever
	}
	return pickOptions{
		categoryName:    strings.TrimSpace(command.Category),
		includeExcluded: command.IncludeExcluded,
		markMode:        markMode,
	}
}

func sortedEnabledKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key, enabled := range values {
		if enabled {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func expandHomePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}
