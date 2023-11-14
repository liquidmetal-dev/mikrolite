package vm

import (
	"fmt"

	ctr "github.com/containerd/containerd"
	"github.com/pterm/pterm"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/mikrolite/mikrolite/adapters/containerd"
	"github.com/mikrolite/mikrolite/adapters/filesystem"
	"github.com/mikrolite/mikrolite/adapters/godisk"
	"github.com/mikrolite/mikrolite/adapters/netlink"
	"github.com/mikrolite/mikrolite/adapters/vm"
	"github.com/mikrolite/mikrolite/core/app"
)

func newRemoveVMCommand(cfg *commonConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a vm",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			vmName := args[0]

			pterm.DefaultSpinner.Start()
			pterm.DefaultSpinner.Info(fmt.Sprintf("🗑️ Deleting VM: %s\n", vmName))

			//TODO: move this to dependency injection
			fsSvc := afero.NewOsFs()
			stateSvc, err := filesystem.NewStateService(vmName, cfg.StateRootPath, fsSvc)
			if err != nil {
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error creating state service: %s\n", err))
				return
			}
			diskSvc := godisk.New(fsSvc)
			netSvc := netlink.New()
			client, err := ctr.New(cfg.SocketPath)
			if err != nil {
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error creating containerd client: %s\n", err))
				return
			}
			imageSvc := containerd.NewImageService(client)
			vmSvc, err := vm.New(cfg.VMProvider, vm.VMProviderProps{
				StateService:       stateSvc,
				DiskSvc:            diskSvc,
				Fs:                 fsSvc,
				FirecrackerBin:     cfg.FirecrackerBin,
				CloudHypervisorBin: cfg.CloudHypervisorBin,
			})
			if err != nil {
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error creating vm provider: %s\n", err))
				return
			}

			owner := fmt.Sprintf("vm-%s", vmName)
			a := app.New(imageSvc, vmSvc, stateSvc, fsSvc, netSvc)
			if err := a.RemoveVM(cmd.Context(), vmName, owner); err != nil {
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error removing vm %s: %s\n", vmName, err))
				return
			}

			pterm.DefaultSpinner.Success(fmt.Sprintf("✅ Succesfully delete VM: %s\n", vmName))
			pterm.DefaultSpinner.Stop()
		},
	}

	return cmd
}
