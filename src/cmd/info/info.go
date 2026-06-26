package info

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"sqill/src/lib/metadata"
	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Display manifest, source, and install metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			entry, err := rt.Store.Get(name)
			if err != nil {
				return err
			}

			manifest, err := metadata.LoadManifest(filepath.Join(rt.SkillsDir, name))
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
