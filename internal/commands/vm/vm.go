package vm

import "github.com/spf13/cobra"

func NewVMCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Create and manage virtual machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newCreateCommandVM())
	cmd.AddCommand(newRemoveVMCommand())

	return cmd
}
