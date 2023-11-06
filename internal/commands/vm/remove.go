package vm

import "github.com/spf13/cobra"

func newRemoveVMCommand(socketPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a vm",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	//TODO: add flags

	return cmd
}
