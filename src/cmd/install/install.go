package install

import (
	"fmt"

	"github.com/spf13/cobra"

	"sqill/src/lib/metadata"
	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "install [<name>]",
		Short: "Install a skill, or all installed skills if no name is given",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			installOne := func(name string) error {
				if err := rt.Inst.Install(name, force); err != nil {
					return err
				}
				if err := metadata.SyncGitignore(rt.SkillsDir); err != nil {
					return fmt.Errorf("sync gitignore: %w", err)
				}
				fmt.Fprintf(out, "Installed %s\n", name)
				return nil
			}

			if len(args) == 1 {
				return installOne(args[0])
			}

			state, err := rt.Store.Load()
			if err != nil {
				return err
			}
			names := metadata.SortedNames(state)
			if len(names) == 0 {
				fmt.Fprintln(out, "No skills to install.")
				return nil
			}

			var failed []string
			for _, name := range names {
				if rt.Store.IsInstalled(name) {
					fmt.Fprintf(out, "%s is already up to date\n", name)
					continue
				}
				if err := installOne(name); err != nil {
					fmt.Fprintf(out, "Failed %s: %v\n", name, err)
					failed = append(failed, name)
				}
			}
			if len(failed) > 0 {
				return fmt.Errorf("%d skill(s) failed to install", len(failed))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing installations")
	return cmd
}
