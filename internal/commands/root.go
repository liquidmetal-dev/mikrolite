package commands

import (
	"github.com/mikrolite/mikrolite/internal/commands/vm"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mikrolite",
		Short: "A CLI to do stuff with microVMs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(vm.NewVMCommand())

	return cmd
}
