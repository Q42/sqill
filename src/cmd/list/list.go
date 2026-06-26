package list

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"sqill/src/lib/metadata"
	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := rt.Store.Load()
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
