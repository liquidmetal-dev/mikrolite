package vm

import (
	"fmt"
	"strconv"

	"github.com/mikrolite/mikrolite/adapters/filesystem"
	"github.com/pterm/pterm"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newListCommandVM(cfg *commonConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List virtual machines",
		Run: func(cmd *cobra.Command, args []string) {
			fsSvc := afero.NewOsFs()
			stateSvc, err := filesystem.NewStateService("", cfg.StateRootPath, fsSvc)
			if err != nil {
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error creating state service: %s\n", err))
				return
			}

			vms, err := stateSvc.ListVMs()
			if err != nil {
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error listing VMs: %s\n", err))
				return
			}

			vmPrintData := [][]string{
				{"Name", "VCPU", "MemoryInMB"},
			}
			for _, vm := range vms {
				vmPrintData = append(vmPrintData, []string{vm.Name, strconv.Itoa(vm.Spec.VCPU), strconv.Itoa(vm.Spec.MemoryInMb)})
			}

			table := pterm.DefaultTable
			table.HasHeader = true

			table.WithData(vmPrintData).Render()
		},
	}

	return cmd
}
