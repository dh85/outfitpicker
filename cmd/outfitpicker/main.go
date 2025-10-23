package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dh85/outfitpicker/internal/app"
	"github.com/dh85/outfitpicker/internal/cli"
	"github.com/dh85/outfitpicker/internal/ui"
	"github.com/dh85/outfitpicker/pkg/config"
	"github.com/dh85/outfitpicker/pkg/version"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		category string
		rootFlag string
		setRoot  string
		quick    bool
		language string
	)

	// Detect locale for help text
	locale := app.DetectLocale()
	i18n := app.NewI18n(locale)

	cmd := &cobra.Command{
		Use:   "outfitpicker [root]",
		Short: getShortDescription(i18n),
		Long:  getLongDescription(i18n),
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// handle --version (alternative to separate command)
			if v, _ := cmd.Flags().GetBool("version"); v {
				fmt.Fprintln(cmd.OutOrStdout(), version.GetVersion())
				return nil
			}

			if setRoot != "" {
				cfg := &config.Config{Root: setRoot}
				if language != "" {
					cfg.Language = language
				}
				if err := config.Save(cfg); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				_ = cli.EnsureCacheAtRoot(setRoot, cmd.OutOrStdout())
				return nil
			}

			// Handle quick mode
			if quick {
				var root string
				if len(args) >= 1 {
					root = args[0]
				} else if rootFlag != "" {
					root = rootFlag
				} else if cfg, err := config.Load(); err == nil && cfg.Root != "" {
					root = cfg.Root
				} else {
					return fmt.Errorf("no root directory specified")
				}

				// Create i18n instance for quick mode
				locale := getLocale(language)
				i18n := app.NewI18n(locale)

				return app.QuickModeRandomWithI18n(root, category, cmd.OutOrStdout(), i18n)
			}

			var root string
			if len(args) >= 1 {
				root = args[0]
			} else if rootFlag != "" {
				root = rootFlag
			} else if cfg, err := config.Load(); err == nil && cfg.Root != "" {
				root = cfg.Root
				// Use enhanced UI for info message
				theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
				uiInstance := ui.NewUI(cmd.OutOrStdout(), theme)
				uiInstance.Info(fmt.Sprintf("using root from config: %s", root))
			} else {
				r, err := cli.FirstRunWizard(cmd.InOrStdin(), cmd.OutOrStdout())
				if err != nil {
					return err
				}
				root = r
			}

			// Create i18n instance
			locale := getLocale(language)
			i18n := app.NewI18n(locale)

			return app.RunWithI18n(root, category, cmd.InOrStdin(), cmd.OutOrStdout(), i18n)
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", getFlagDescription(i18n, "category_flag"))
	cmd.Flags().StringVar(&rootFlag, "root", "", getFlagDescription(i18n, "root_flag"))
	cmd.Flags().StringVar(&setRoot, "set-root", "", getFlagDescription(i18n, "set_root_flag"))
	cmd.Flags().BoolVarP(&quick, "quick", "q", false, getFlagDescription(i18n, "quick_flag"))
	cmd.Flags().StringVarP(&language, "language", "l", "", getFlagDescription(i18n, "language_flag"))
	cmd.Flags().BoolP("version", "v", false, getFlagDescription(i18n, "version_flag"))

	// subcommands
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newCompletionCmd(cmd))

	return cmd
}

// shouldUseColors determines if colors should be used based on environment
func shouldUseColors() bool {
	// Check if output is a terminal and colors are supported
	if term := os.Getenv("TERM"); term == "dumb" || term == "" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

// getLocale determines the locale to use based on flag, config, and environment
func getLocale(languageFlag string) string {
	// 1. Command line flag takes precedence
	if languageFlag != "" {
		return languageFlag
	}

	// 2. Check config file
	if cfg, err := config.Load(); err == nil && cfg.Language != "" {
		return cfg.Language
	}

	// 3. Auto-detect from environment
	return app.DetectLocale()
}

// Helper functions for i18n command descriptions
func getShortDescription(i18n *app.I18n) string {
	switch i18n.GetLocale() {
	case "es":
		return "Selecciona archivos de outfit desde carpetas de categorías"
	case "fr":
		return "Sélectionner des fichiers de tenues depuis des dossiers de catégories"
	case "de":
		return "Outfit-Dateien aus Kategorieordnern auswählen"
	case "it":
		return "Seleziona file di outfit dalle cartelle delle categorie"
	case "pt":
		return "Selecionar arquivos de roupas de pastas de categorias"
	case "nl":
		return "Selecteer outfit bestanden uit categorie mappen"
	case "ru":
		return "Выбрать файлы нарядов из папок категорий"
	case "ja":
		return "カテゴリフォルダから服装ファイルを選択"
	case "zh":
		return "从类别文件夹中选择服装文件"
	default:
		return "Select outfit files from category folders"
	}
}

func getLongDescription(i18n *app.I18n) string {
	switch i18n.GetLocale() {
	case "es":
		return "CLI interactivo para elegir outfits desde carpetas de categorías, con rotación por categoría almacenada en tu directorio sincronizado."
	case "fr":
		return "CLI interactif pour choisir des tenues depuis des dossiers de catégories, avec une rotation par catégorie mise en cache dans votre répertoire synchronisé."
	case "de":
		return "Interaktive CLI zur Auswahl von Outfits aus Kategorieordnern, mit kategoriebasierter Rotation im synchronisierten Verzeichnis."
	case "it":
		return "CLI interattiva per scegliere outfit dalle cartelle delle categorie, con rotazione per categoria memorizzata nella directory sincronizzata."
	case "pt":
		return "CLI interativo para escolher roupas de pastas de categorias, com rotação por categoria armazenada em seu diretório sincronizado."
	case "nl":
		return "Interactieve CLI om outfits te kiezen uit categoriemappen, met rotatie per categorie opgeslagen in je gesynchroniseerde map."
	case "ru":
		return "Интерактивный CLI для выбора нарядов из папок категорий, с ротацией по категориям в синхронизированном каталоге."
	case "ja":
		return "カテゴリフォルダから服装を選択するためのインタラクティブCLI、同期されたルートでカテゴリ別ローテーションをキャッシュ。"
	case "zh":
		return "交互式CLI，从类别文件夹中选择服装，在同步根目录中缓存按类别轮换。"
	default:
		return "Interactive CLI to pick outfits from category folders, with a per-category rotation cached in your synced root."
	}
}

func getFlagDescription(i18n *app.I18n, key string) string {
	locale := i18n.GetLocale()

	// Language-specific descriptions
	switch locale {
	case "es":
		switch key {
		case "category_flag":
			return "carpeta de categoría a usar (ej: Playa, Latex, General)"
		case "language_flag":
			return "idioma para la interfaz (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, etc.)"
		}
	case "fr":
		switch key {
		case "language_flag":
			return "langue pour l'interface (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, etc.)"
		}
	case "de":
		switch key {
		case "language_flag":
			return "Sprache für die Benutzeroberfläche (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, etc.)"
		}
	case "ru":
		switch key {
		case "language_flag":
			return "язык для интерфейса (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, и т.д.)"
		}
	case "ja":
		switch key {
		case "language_flag":
			return "インターフェースの言語 (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, など)"
		}
	case "zh":
		switch key {
		case "language_flag":
			return "界面语言 (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, 等)"
		}
	}

	// English defaults
	switch key {
	case "category_flag":
		return "category folder to use (e.g., Beach, Latex, General)"
	case "root_flag":
		return "root 'Outfits' directory (overrides config)"
	case "set_root_flag":
		return "persist this path as the default root in user config, then exit"
	case "quick_flag":
		return "quick mode - instantly select random outfit without menus"
	case "language_flag":
		return "language for interface (en, es, fr, de, it, pt, nl, ru, ja, zh, ko, ar, hi, etc.)"
	case "version_flag":
		return "print version and exit"
	default:
		return ""
	}
}

func getConfigShortDescription(i18n *app.I18n) string {
	if i18n.GetLocale() == "es" {
		return "Gestionar configuración"
	}
	return "Manage configuration"
}

func getConfigShowDescription(i18n *app.I18n) string {
	if i18n.GetLocale() == "es" {
		return "Mostrar archivo de configuración y ruta raíz"
	}
	return "Show config file and root path"
}

func getConfigSetRootDescription(i18n *app.I18n) string {
	if i18n.GetLocale() == "es" {
		return "Establecer directorio raíz predeterminado"
	}
	return "Set default root directory"
}

func getConfigResetDescription(i18n *app.I18n) string {
	if i18n.GetLocale() == "es" {
		return "Eliminar el archivo de configuración (activa el asistente de primera ejecución la próxima vez)"
	}
	return "Delete the config file (triggers first-run wizard next time)"
}

func getCompletionDescription(i18n *app.I18n) string {
	if i18n.GetLocale() == "es" {
		return "Generar scripts de autocompletado para shell"
	}
	return "Generate shell completion scripts"
}

func newConfigCmd() *cobra.Command {
	locale := app.DetectLocale()
	i18n := app.NewI18n(locale)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: getConfigShortDescription(i18n),
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: getConfigShowDescription(i18n),
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := config.Path()
			if err != nil {
				return fmt.Errorf("failed to locate config file: %w", err)
			}
			cfg, err := config.Load()
			switch {
			case err == os.ErrNotExist:
				fmt.Fprintf(cmd.OutOrStdout(), "config file: %s (not found)\n", p)
				return nil
			case err != nil:
				return fmt.Errorf("failed to load config: %w", err)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "config file: %s\nroot: %s\n", p, cfg.Root)
				return nil
			}
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "set-root <path>",
		Short: getConfigSetRootDescription(i18n),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Save(&config.Config{Root: args[0]}); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			// Use enhanced UI for success message
			theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
			uiInstance := ui.NewUI(cmd.OutOrStdout(), theme)
			uiInstance.Success("saved default root")
			_ = cli.EnsureCacheAtRoot(args[0], cmd.OutOrStdout())
			return nil
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: getConfigResetDescription(i18n),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Delete(); err != nil {
				return fmt.Errorf("failed to reset config: %w", err)
			}
			// Use enhanced UI for success message
			theme := ui.Theme{UseColors: shouldUseColors(), UseEmojis: true, Compact: true}
			uiInstance := ui.NewUI(cmd.OutOrStdout(), theme)
			uiInstance.Success("config reset")
			return nil
		},
	})

	return configCmd
}

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	locale := app.DetectLocale()
	i18n := app.NewI18n(locale)

	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: getCompletionDescription(i18n),
		Args:  cobra.ExactValidArgs(1),
		ValidArgs: []string{
			"bash", "zsh", "fish", "powershell",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unknown shell %q", args[0])
			}
		},
	}
}
