package vm

import (
	"fmt"

	ctr "github.com/containerd/containerd"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/mikrolite/mikrolite/adapters/containerd"
	"github.com/mikrolite/mikrolite/adapters/vm"
	"github.com/mikrolite/mikrolite/core/app"
	"github.com/mikrolite/mikrolite/core/domain"
)

func newCreateCommandVM(socketPath *string) *cobra.Command {
	input := struct {
		Name                      string
		VCPU                      int
		MemoryInMb                int
		RootVolumeImage           string
		KernelVolumeImage         string
		KernelVolumeImageFilename string
		KernelHostPath            string
	}{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {

			spec := &domain.VMSpec{
				VCPU:       input.VCPU,
				MemoryInMb: input.MemoryInMb,
				Kernel: domain.Kernel{
					Source: domain.KernelSource{},
				},
				RootVolume: domain.Volume{
					Name: "root",
					Source: domain.VolumeSource{
						Container: &domain.ContainerVolumeSource{
							Image: input.RootVolumeImage,
						},
					},
				},
			}
			if input.KernelVolumeImage != "" {
				spec.Kernel.Source.Container = &domain.ContainerKernelSource{
					Image:    input.KernelVolumeImage,
					Filename: input.KernelVolumeImageFilename,
				}
			}
			if input.KernelHostPath != "" {
				spec.Kernel.Source.HostPath = &domain.HostPathKernelSource{
					Path: input.KernelHostPath,
				}
			}

			client, err := ctr.New(*socketPath)
			if err != nil {
				return fmt.Errorf("creating containerd client: %w", err)
			}
			imageSvc := containerd.NewImageService(client)
			vmSvc, err := vm.New("firecracker")
			if err != nil {
				return fmt.Errorf("creating firecracker vm provider: %w", err)
			}
			fsSvc := afero.NewOsFs()

			a := app.New(imageSvc, vmSvc, fsSvc)
			vm, err := a.CreateVM(cmd.Context(), input.Name, spec)
			if err != nil {
				return fmt.Errorf("failed creating vm: %w", err)
			}
			fmt.Println(vm)

			return nil
		},
	}

	cmd.Flags().StringVarP(&input.Name, "name", "n", "", "The name of the vm")
	cmd.Flags().IntVarP(&input.VCPU, "cpu", "c", 2, "The number of virtual cpus")
	cmd.Flags().IntVarP(&input.MemoryInMb, "memory", "m", 2048, "The amount of memory for the vm")
	cmd.Flags().StringVar(&input.RootVolumeImage, "root-image", "", "The container to use for the root volume")
	cmd.Flags().StringVar(&input.KernelVolumeImage, "kernel-image", "", "The container to use for the kernel")
	cmd.Flags().StringVar(&input.KernelVolumeImageFilename, "kernel-image-filename", "vmlinux", "The name of the kernel file in the image")
	cmd.Flags().StringVar(&input.KernelHostPath, "kernel-path", "", "The path to a kernel file on the host")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("root-image")
	cmd.MarkFlagsMutuallyExclusive("kernel-image", "kernel-path")

	return cmd
}