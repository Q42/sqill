package untrack

import (
	"fmt"

	"github.com/spf13/cobra"

	"sqill/src/lib/metadata"
	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "untrack <name>",
		Short: "Exclude an installed skill's directory from git",
		Long:  "Adds the skill's directory to .agents/skills/.gitignore so it is not committed. Idempotent: untracking a skill that is not currently tracked is a no-op.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			out := cmd.OutOrStdout()
			if !rt.Store.IsTracked(name) {
				fmt.Fprintf(out, "%s is already untracked\n", name)
				return nil
			}
			if err := rt.Store.Untrack(name); err != nil {
				return err
			}
			if err := metadata.SyncGitignore(rt.SkillsDir); err != nil {
				return fmt.Errorf("sync gitignore: %w", err)
			}
			fmt.Fprintf(out, "Untracked %s\n", name)
			return nil
		},
	}
}
