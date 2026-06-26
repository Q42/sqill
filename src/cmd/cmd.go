package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"sqill/src/cmd/info"
	initcmd "sqill/src/cmd/init"
	"sqill/src/cmd/install"
	"sqill/src/cmd/list"
	"sqill/src/cmd/remove"
	"sqill/src/cmd/search"
	"sqill/src/cmd/update"
	"sqill/src/lib/buildinfo"
	"sqill/src/lib/metadata"
	"sqill/src/lib/runtime"
)

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

	rt := &runtime.Runtime{}

	for _, sub := range []*cobra.Command{
		initcmd.NewCmd(),
		install.NewCmd(rt),
		remove.NewCmd(rt),
		update.NewCmd(rt),
		list.NewCmd(rt),
		search.NewCmd(rt),
		info.NewCmd(rt),
	} {
		r.AddCommand(sub)
	}

	r.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		absSkillsDir, err := filepath.Abs(skillsDir)
		if err != nil {
			return fmt.Errorf("resolve skills dir: %w", err)
		}

		if !requiresInitializedState(cmd) {
			rt.SkillsDir = absSkillsDir
			return nil
		}

		statePath := filepath.Join(absSkillsDir, metadata.StateFileName)
		if _, err := os.Stat(statePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%s not found. Run `sqill init` to initialize", statePath)
			}
			return fmt.Errorf("check state: %w", err)
		}

		full, err := runtime.New(absSkillsDir)
		if err != nil {
			return err
		}
		*rt = *full
		return nil
	}

	return r
}

func requiresInitializedState(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "init", "help", "completion":
		return false
	}
	return true
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}