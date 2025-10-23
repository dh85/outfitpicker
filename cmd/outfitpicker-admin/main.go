package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dh85/outfitpicker/internal/storage"
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
		rootOverride string
		showVersion  bool
	)

	cmd := &cobra.Command{
		Use:   "outfitpicker-admin",
		Short: "Admin utilities for outfitpicker",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// handle --version early for any command
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), version.GetVersion())
				os.Exit(0)
			}
			return nil
		},
	}

	// Global flags
	cmd.PersistentFlags().StringVar(&rootOverride, "root", "", "override the saved root 'Outfits' directory for this command")
	cmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "print version and exit")

	// Subcommands
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newCacheCmd(&rootOverride))

	return cmd
}

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show config file and root path",
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
		Short: "Set default root directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Save(&config.Config{Root: args[0]}); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "✅ saved default root")
			return nil
		},
	})

	configCmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Delete the config file (triggers first-run wizard next time)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := config.Delete(); err != nil {
				return fmt.Errorf("failed to reset config: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "✅ config reset")
			return nil
		},
	})

	return configCmd
}

func newCacheCmd(rootOverride *string) *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Inspect and manage selection cache",
	}

	cacheCmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show cache file path and per-category counts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(*rootOverride)
			if err != nil {
				return err
			}
			mgr, err := storage.NewManager(root)
			if err != nil {
				return fmt.Errorf("failed to init cache manager: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "cache: %s\n", mgr.Path())

			cm := mgr.Load()
			if len(cm) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(empty)")
				return nil
			}
			// Sort by category base name for nice output
			type row struct {
				path, name string
				n          int
			}
			var rows []row
			for k, v := range cm {
				rows = append(rows, row{
					path: k,
					name: filepath.Base(k),
					n:    len(v),
				})
			}
			sort.Slice(rows, func(i, j int) bool { return strings.ToLower(rows[i].name) < strings.ToLower(rows[j].name) })
			for _, r := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s  %d selected\n", r.name, r.n)
			}
			return nil
		},
	})

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear cache for a category or all categories",
		Long:  "Without flags, clears a single category by name.\nUse --all to clear all categories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(*rootOverride)
			if err != nil {
				return err
			}
			mgr, err := storage.NewManager(root)
			if err != nil {
				return fmt.Errorf("failed to init cache manager: %w", err)
			}

			all, _ := cmd.Flags().GetBool("all")
			if all {
				cm := mgr.Load()
				if len(cm) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "cache already empty")
					return nil
				}
				// Clear every category key present in the cache file
				keys := make([]string, 0, len(cm))
				for k := range cm {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					mgr.Clear(k)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "✅ cleared cache for all categories")
				return nil
			}

			if len(args) < 1 {
				return fmt.Errorf("provide a category name, or use --all")
			}
			category := args[0]

			// We don't need to read the FS; cache keys are full category paths.
			// Find the key whose base matches the provided category (case-insensitive).
			cm := mgr.Load()
			var matched string
			for k := range cm {
				if strings.EqualFold(filepath.Base(k), category) {
					matched = k
					break
				}
			}
			if matched == "" {
				// Helpful message: list available categories (from cache)
				var names []string
				for k := range cm {
					names = append(names, filepath.Base(k))
				}
				sort.Strings(names)
				if len(names) == 0 {
					return fmt.Errorf("no categories found in cache; try 'outfitpicker-admin cache show' first")
				}
				return fmt.Errorf("category %q not found in cache; available: %s", category, strings.Join(names, ", "))
			}

			mgr.Clear(matched)
			fmt.Fprintf(cmd.OutOrStdout(), "✅ cleared cache for %q\n", filepath.Base(matched))
			return nil
		},
	}
	clearCmd.Flags().Bool("all", false, "clear all categories in the cache")
	cacheCmd.AddCommand(clearCmd)

	return cacheCmd
}

func resolveRoot(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	cfg, err := config.Load()
	if err == os.ErrNotExist {
		return "", fmt.Errorf("no saved root; run 'outfitpicker --set-root <path>' or pass --root")
	}
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.Root == "" {
		return "", fmt.Errorf("config has empty root; run 'outfitpicker config set-root <path>'")
	}
	return cfg.Root, nil
}
