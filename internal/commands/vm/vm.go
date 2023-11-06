package vm

import "github.com/spf13/cobra"

func NewVMCommand() *cobra.Command {
	socketPath := ""

	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Create and manage virtual machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringVar(&socketPath, "socket-path", "/run/containerd/containerd.sock", "the path to the containerd socket")

	cmd.AddCommand(newCreateCommandVM(&socketPath))
	cmd.AddCommand(newRemoveVMCommand(&socketPath))

	return cmd
}
