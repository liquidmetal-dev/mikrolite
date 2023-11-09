package vm

import (
	"fmt"

	ctr "github.com/containerd/containerd"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			vmName := args[0]

			//TODO: move this to dependency injection
			fsSvc := afero.NewOsFs()
			stateSvc, err := filesystem.NewStateService(vmName, cfg.StateRootPath, fsSvc)
			if err != nil {
				return fmt.Errorf("creating state service: %w", err)
			}
			diskSvc := godisk.New(fsSvc)
			netSvc := netlink.New()
			client, err := ctr.New(cfg.SocketPath)
			if err != nil {
				return fmt.Errorf("creating containerd client: %w", err)
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
				return fmt.Errorf("creating vm provider %s: %w", cfg.VMProvider, err)
			}

			owner := fmt.Sprintf("vm-%s", vmName)
			a := app.New(imageSvc, vmSvc, stateSvc, fsSvc, netSvc)
			if err := a.RemoveVM(cmd.Context(), vmName, owner); err != nil {
				return fmt.Errorf("failed creating vm: %w", err)
			}

			return nil
		},
	}

	return cmd
}
