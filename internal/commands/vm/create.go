package vm

import (
	"errors"
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
	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/mikrolite/mikrolite/defaults"
)

func newCreateCommandVM(cfg *commonConfig) *cobra.Command {
	input := struct {
		Name              string
		VCPU              int
		MemoryInMb        int
		RootVolumeImage   string
		KernelVolumeImage string
		KernelFilename    string
		KernelHostPath    string
		BridgeName        string
		StaticIP          string
		StaticGatewayIP   string
		SSHKeyFile        string
		SnapshotterVolume string
		SnapshotterKernel string
	}{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {

			pterm.DefaultSpinner.Start()
			pterm.DefaultSpinner.Info(fmt.Sprintf("🚀 Creating VM: %s\n", input.Name))

			spec := &domain.VMSpec{
				VCPU:       input.VCPU,
				MemoryInMb: input.MemoryInMb,
				Kernel: domain.Kernel{
					Source: domain.KernelSource{
						Filename: input.KernelFilename,
					},
				},
				RootVolume: domain.Volume{
					Name: "root",
					Source: domain.VolumeSource{
						Container: &domain.ContainerVolumeSource{
							Image: input.RootVolumeImage,
						},
					},
				},
				NetworkConfiguration: domain.NetworkConfiguration{
					BridgeName: input.BridgeName,
					Interfaces: map[string]domain.NetwortInterface{},
				},
			}
			if input.KernelVolumeImage != "" {
				spec.Kernel.Source.Container = &domain.ContainerKernelSource{
					Image: input.KernelVolumeImage,
				}
			}
			if input.KernelHostPath != "" {
				spec.Kernel.Source.HostPath = &domain.HostPathKernelSource{
					Path: input.KernelHostPath,
				}
			}
			netInt := domain.NetwortInterface{
				GuestDeviceName:       "eth0",
				AllowMetadataRequests: false,
				AttachToBridge:        true,
			}

			if input.StaticIP != "" {
				netInt.StaticIPv4Address = &domain.StaticIPv4Address{
					Address: input.StaticIP,
				}

				if input.StaticGatewayIP != "" {
					netInt.StaticIPv4Address.Gateway = &input.StaticGatewayIP
				}
			}
			spec.NetworkConfiguration.Interfaces["eth0"] = netInt

			if input.SSHKeyFile != "" {
				spec.Bootstrap = &domain.Bootstrap{
					SSHKey: input.SSHKeyFile,
				}
			}

			//TODO: move this to dependency injection
			fsSvc := afero.NewOsFs()
			stateSvc, err := filesystem.NewStateService(input.Name, cfg.StateRootPath, fsSvc)
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
				pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error creating vm provider %s: %s\n", cfg.VMProvider, err))
				return
			}

			owner := fmt.Sprintf("vm-%s", input.Name)
			a := app.New(imageSvc, vmSvc, stateSvc, fsSvc, netSvc)
			vm, err := a.CreateVM(cmd.Context(), ports.CreateVMInput{
				Name:              input.Name,
				Owner:             owner,
				Spec:              spec,
				SnapshotterKernal: input.SnapshotterKernel,
				SnapshotterVolume: input.SnapshotterVolume,
			})
			if err != nil {
				switch {
				case errors.Is(err, app.ErrVMAlreadyExists):
					pterm.DefaultSpinner.Warning(fmt.Sprintf("VM with name %s already exists\n", input.Name))
					return
				default:
					pterm.DefaultSpinner.Fail(fmt.Sprintf("❌ Error creating vm %s: %s\n", input.Name, err))
					return
				}
			}

			pterm.DefaultSpinner.Success(fmt.Sprintf("✅ Succesfully created VM: %s (%s)\n", input.Name, vm.Status.IP))
			pterm.DefaultSpinner.Stop()
		},
	}

	cmd.Flags().StringVarP(&input.Name, "name", "n", "", "The name of the vm")
	cmd.Flags().IntVarP(&input.VCPU, "cpu", "c", 2, "The number of virtual cpus")
	cmd.Flags().IntVarP(&input.MemoryInMb, "memory", "m", 2048, "The amount of memory for the vm")
	cmd.Flags().StringVar(&input.RootVolumeImage, "root-image", "", "The container to use for the root volume")
	cmd.Flags().StringVar(&input.KernelVolumeImage, "kernel-image", "", "The container to use for the kernel")
	cmd.Flags().StringVar(&input.KernelHostPath, "kernel-path", "", "The path to a kernel file on the host")
	cmd.Flags().StringVar(&input.KernelFilename, "kernel-filename", "vmlinux", "The name of the kernel file in the image or in the hostpath")
	cmd.Flags().StringVar(&input.BridgeName, "network-bridge", defaults.SharedBridgeName, "The name of the bridge to attach the vm to")
	cmd.Flags().StringVar(&input.StaticIP, "static-ip", "", "A static IPV4 address (as a CIDR) to assign to the VM. If ommitted DHCP will be used")
	cmd.Flags().StringVar(&input.StaticGatewayIP, "static-gateway-ip", "", "A gateway (as a CIDR) to use with the static IP")
	cmd.Flags().StringVar(&input.SSHKeyFile, "ssh-key", "", "A SSH public key to use as an authorized key")
	cmd.Flags().StringVar(&input.SnapshotterVolume, "snapshotter-volume", "devmapper", "The containerd snapshotter to use for volumes")
	cmd.Flags().StringVar(&input.SnapshotterKernel, "snapshotter-kernel", "native", "The containerd snapshotter to use for the kernel")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("root-image")
	cmd.MarkFlagsMutuallyExclusive("kernel-image", "kernel-path")

	return cmd
}
