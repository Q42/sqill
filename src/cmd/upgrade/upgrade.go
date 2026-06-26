package upgrade

import (
	"fmt"

	"github.com/spf13/cobra"

	"sqill/src/lib/buildinfo"
	"sqill/src/lib/upgrader"
)

func NewCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the sqill binary to the latest release",
		Long:  "Downloads the latest sqill release for your OS/arch and atomically replaces the running binary.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			result, err := upgrader.Run(upgrader.Options{
				Version: buildinfo.Version,
				Force:   force,
			})
			if err != nil {
				return err
			}
			if result.Skipped {
				fmt.Fprintf(out, "Already on latest version (%s).\n", result.To)
				return nil
			}
			fmt.Fprintf(out, "Upgraded sqill: %s -> %s\n", result.From, result.To)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Reinstall even if already on the latest version")
	return cmd
}
