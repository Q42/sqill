package track

import (
	"fmt"

	"github.com/spf13/cobra"

	"sqill/src/lib/metadata"
	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "track <name>",
		Short: "Include an installed skill's directory in git",
		Long:  "Removes the skill's directory from .agents/skills/.gitignore so it is committed to version control. The skill must already be installed.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := rt.Store.Track(name); err != nil {
				return err
			}
			if err := metadata.SyncGitignore(rt.SkillsDir); err != nil {
				return fmt.Errorf("sync gitignore: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Tracking %s\n", name)
			return nil
		},
	}
}
