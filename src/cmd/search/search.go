package search

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Find matching skills in the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := rt.Reg.Search(args[0])
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
