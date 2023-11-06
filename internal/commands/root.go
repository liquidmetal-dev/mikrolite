package commands

import "github.com/spf13/cobra"

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mikrolite",
		Short: "A CLI to do stuff with microVMs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	//TODO: add subcommands

	return cmd
}
