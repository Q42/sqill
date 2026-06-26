package update

import (
	"fmt"

	"github.com/spf13/cobra"

	"sqill/src/lib/runtime"
)

func NewCmd(rt *runtime.Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "update <name>",
		Short: "Fetch latest version and replace atomically",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := rt.Inst.Update(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", args[0])
			return nil
		},
	}
}
