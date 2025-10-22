package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dh85/outfitpicker/internal/app"
	"github.com/dh85/outfitpicker/internal/cli"
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
	)

	cmd := &cobra.Command{
		Use:   "outfitpicker [root]",
		Short: "Select outfit files from category folders",
		Long:  "Interactive CLI to pick outfits from category folders, with a per-category rotation cached in your synced root.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// handle --version (alternative to separate command)
			if v, _ := cmd.Flags().GetBool("version"); v {
				fmt.Fprintln(cmd.OutOrStdout(), version.Version)
				return nil
			}

			if setRoot != "" {
				if err := config.Save(&config.Config{Root: setRoot}); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
				_ = cli.EnsureCacheAtRoot(setRoot, cmd.OutOrStdout())
				return nil
			}

			var root string
			if len(args) >= 1 {
				root = args[0]
			} else if rootFlag != "" {
				root = rootFlag
			} else if cfg, err := config.Load(); err == nil && cfg.Root != "" {
				root = cfg.Root
				fmt.Fprintf(cmd.OutOrStdout(), "ℹ️ using root from config: %s\n", root)
			} else {
				r, err := cli.FirstRunWizard(cmd.InOrStdin(), cmd.OutOrStdout())
				if err != nil {
					return err
				}
				root = r
			}

			return app.Run(root, category, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "category folder to use (e.g., Beach, Latex, General)")
	cmd.Flags().StringVar(&rootFlag, "root", "", "root 'Outfits' directory (overrides config)")
	cmd.Flags().StringVar(&setRoot, "set-root", "", "persist this path as the default root in user config, then exit")
	cmd.Flags().BoolP("version", "v", false, "print version and exit")

	// subcommands
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newCompletionCmd(cmd))

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
			_ = cli.EnsureCacheAtRoot(args[0], cmd.OutOrStdout())
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

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
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
