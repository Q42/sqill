package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"sqill/internal/installer"
	"sqill/internal/metadata"
	"sqill/internal/registry"
)

type runtime struct {
	skillsDir string
	store     metadata.Store
	inst      *installer.Installer
	reg       registry.Provider
}

func newRuntime(skillsDir string) (*runtime, error) {
	abs, err := filepath.Abs(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("resolve skills dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create skills dir: %w", err)
	}
	store, err := metadata.NewFileStore(abs)
	if err != nil {
		return nil, err
	}
	reg := registry.NewHardcoded()
	return &runtime{
		skillsDir: abs,
		store:     store,
		inst:      installer.New(reg, store, abs),
		reg:       reg,
	}, nil
}

func NewRoot() *cobra.Command {
	r := &cobra.Command{
		Use:           "sqill",
		Short:         "Agent skill registry CLI",
		Long:          "Sqill installs, updates, removes, and discovers agent skills.",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Usage: sqill <command> [flags]")
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'sqill --help' for the full command list.")
			return fmt.Errorf("missing command")
		},
	}

	var skillsDir string
	r.PersistentFlags().StringVar(&skillsDir, "skills-dir", ".agents/skills", "Directory containing installed skills and metadata")

	rt := &runtime{}

	for _, sub := range []*cobra.Command{
		newSetupCmd(),
		newInstallCmd(rt),
		newRemoveCmd(rt),
		newUpdateCmd(rt),
		newListCmd(rt),
		newSearchCmd(rt),
		newInfoCmd(rt),
	} {
		r.AddCommand(sub)
	}

	r.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		absSkillsDir, err := filepath.Abs(skillsDir)
		if err != nil {
			return fmt.Errorf("resolve skills dir: %w", err)
		}

		if !requiresInitializedState(cmd) {
			rt.skillsDir = absSkillsDir
			return nil
		}

		statePath := filepath.Join(absSkillsDir, metadata.StateFileName)
		if _, err := os.Stat(statePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found. Run `sqill setup` to initialize", statePath)
			}
			return fmt.Errorf("check state: %w", err)
		}

		rt2, err := newRuntime(absSkillsDir)
		if err != nil {
			return err
		}
		*rt = *rt2
		return nil
	}

	return r
}

func requiresInitializedState(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "setup", "help", "completion":
		return false
	}
	return true
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func newInstallCmd(rt *runtime) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "install <name>",
		Short: "Install a skill from the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := rt.inst.Install(args[0], force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing installation")
	return cmd
}

func newRemoveCmd(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Delete an installed skill and its metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := rt.inst.Remove(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", args[0])
			return nil
		},
	}
}

func newUpdateCmd(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "update <name>",
		Short: "Fetch latest version and replace atomically",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := rt.inst.Update(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", args[0])
			return nil
		},
	}
}

func newListCmd(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := rt.store.Load()
			if err != nil {
				return err
			}
			names := metadata.SortedNames(state)
			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No skills installed.")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tVERSION\tINSTALLED")
			for _, n := range names {
				entry := state.Installed[n]
				fmt.Fprintf(tw, "%s\t%s\t%s\n", n, entry.Version, entry.InstalledAt)
			}
			return tw.Flush()
		},
	}
}

func newSearchCmd(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Find matching skills in the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := rt.reg.Search(args[0])
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No matches.")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tSOURCE\tDESCRIPTION")
			for _, e := range entries {
				fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, e.Source, e.Description)
			}
			return tw.Flush()
		},
	}
}

func newInfoCmd(rt *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Display manifest, source, and install metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			entry, err := rt.store.Get(name)
			if err != nil {
				return err
			}

			manifest, err := metadata.LoadManifest(filepath.Join(rt.skillsDir, name))
			if err != nil {
				return fmt.Errorf("load manifest: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Name:        %s\n", manifest.Name)
			fmt.Fprintf(out, "Version:     %s\n", manifest.Version)
			if manifest.Description != "" {
				fmt.Fprintf(out, "Description: %s\n", manifest.Description)
			}
			fmt.Fprintf(out, "Source:      %s\n", entry.Source)
			fmt.Fprintf(out, "Installed:   %s\n", entry.InstalledAt)
			return nil
		},
	}
}
