package commands

import (
	"fmt"

	"github.com/mikrolite/mikrolite/internal/commands/vm"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mikrolite",
		Short: "A CLI to do stuff with microVMs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println()
			pterm.DefaultBigText.WithLetters(
				putils.LettersFromStringWithStyle("Mikro", pterm.NewStyle(pterm.FgLightGreen)),
				putils.LettersFromStringWithStyle("lite", pterm.NewStyle(pterm.FgLightBlue))).
				Render()
			fmt.Println()
			return cmd.Help()
		},
	}

	cmd.AddCommand(vm.NewVMCommand())

	return cmd
}
