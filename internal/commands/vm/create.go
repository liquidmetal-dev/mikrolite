package vm

import (
	"github.com/spf13/cobra"
)

func newCreateCommandVM() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}

	//TODO: add flags

	return cmd
}
