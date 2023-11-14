package firecracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikrolite/mikrolite/adapters/vm/shared"
	"github.com/mikrolite/mikrolite/cloudinit"

	sdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/defaults"
)

// Create will create a new vm.
func (f *Provider) Create(ctx context.Context, vm *domain.VM) (string, error) {
	socketPath := f.socketPath()
	kernelPath := filepath.Join(vm.Status.KernelMount.Location, vm.Spec.Kernel.Source.Filename)
	//networkCfgPath := fmt.Sprintf("%s/fcnet.conflist", f.ss.Root())
	if len(vm.Spec.Kernel.CmdLine) == 0 {
		vm.Spec.Kernel.CmdLine = defaultKernelCmdLine()
	}

	if err := f.ensureLogPath(); err != nil {
		return "", fmt.Errorf("ensuring log file is created: %w", err)
	}

	stdOutFile, err := f.fs.OpenFile(f.ss.StdoutPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaults.DataFilePerm)
	if err != nil {
		return "", fmt.Errorf("opening stdout file %s: %w", f.ss.StdoutPath(), err)
	}

	stdErrFile, err := f.fs.OpenFile(f.ss.StderrPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaults.DataFilePerm)
	if err != nil {
		return "", fmt.Errorf("opening sterr file %s: %w", f.ss.StderrPath(), err)
	}

	metadataFile := ""
	if len(vm.Status.Metadata) > 0 {
		metadataFile, err = f.saveMetadata(vm)
		if err != nil {
			return "", fmt.Errorf("saving metadata to file: %w", err)
		}

		vm.Spec.Kernel.CmdLine["ds"] = "nocloud-net;s=http://169.254.169.254/latest/"
		vm.Spec.Kernel.CmdLine[cloudinit.NetworkConfigDataKey] = vm.Status.Metadata[cloudinit.NetworkConfigDataKey]
	}

	//f.writeNetworkConfig(networkCfgPath, "fcnet")

	cfg := sdk.Config{
		VMID: vm.Name,
		//NetNS:           vm.Status.NetworkNamespace,
		SocketPath:      socketPath,
		KernelImagePath: kernelPath,
		KernelArgs:      shared.FormatKernelCmdLine(vm.Spec.Kernel.CmdLine),
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  intTo64Ptr(vm.Spec.VCPU),
			MemSizeMib: intTo64Ptr(vm.Spec.MemoryInMb),
			Smt:        boolPtr(true),
		},
		Drives:   []models.Drive{},
		LogPath:  f.ss.LogPath(), //TODO: should we change to use a fifo?
		LogLevel: "Debug",
	}

	// if len(vm.Status.Metadata) > 0 {
	// 	cloudInitFile, err := shared.CreateCloudInitImage(ctx, false, vm, f.ss, f.ds)
	// 	if err != nil {
	// 		return "", fmt.Errorf("creating cloud-init disk image: %w", err)
	// 	}
	// 	cfg.Drives = append(cfg.Drives, models.Drive{
	// 		DriveID:      strPtr(cloudinit.VolumeName),
	// 		IsReadOnly:   boolPtr(true),
	// 		IsRootDevice: boolPtr(false),
	// 		PathOnHost:   strPtr(cloudInitFile),
	// 	})
	// }

	for id, mount := range vm.Status.VolumeMounts {
		isRoot := id == "root"
		drive := models.Drive{
			DriveID:      strPtr(id),
			IsRootDevice: &isRoot,
			IsReadOnly:   boolPtr(false),
			PathOnHost:   &mount.Location,
		}
		cfg.Drives = append(cfg.Drives, drive)
	}

	// cfg.NetworkInterfaces = sdk.NetworkInterfaces{
	// 	{
	// 		CNIConfiguration: &sdk.CNIConfiguration{
	// 			NetworkName: "fcnet",
	// 			IfName:      "veth0",
	// 			ConfDir:     f.ss.Root(),              //TODO: cni conf dir
	// 			BinPath:     []string{"/opt/cni/bin"}, //TODO: path to cni bins
	// 			VMIfName:    "eth0",
	// 		},
	// 		AllowMMDS: true,
	// 	},
	// }
	cfg.NetworkInterfaces = sdk.NetworkInterfaces{}
	for name, netInt := range vm.Spec.NetworkConfiguration.Interfaces {
		status, ok := vm.Status.NetworkStatus[name]
		if !ok {
			return "", fmt.Errorf("failed to get network status for %s: %w", name, err)
		}

		netInt := sdk.NetworkInterface{
			StaticConfiguration: &sdk.StaticNetworkConfiguration{
				MacAddress:  status.GuestMAC,
				HostDevName: status.HostDeviveName,
			},
			AllowMMDS: netInt.AllowMetadataRequests,
		}

		cfg.NetworkInterfaces = append(cfg.NetworkInterfaces, netInt)
	}
	cfg.MmdsVersion = sdk.MMDSv1

	args := []string{}
	if metadataFile != "" {
		args = append(args, "--metadata", metadataFile)
	}

	//TODO: this needs to be an optional arg for the path
	cmd := sdk.VMCommandBuilder{}.
		WithSocketPath(socketPath).
		WithBin(f.binaryPath).
		WithStderr(stdErrFile).
		WithStdout(stdOutFile).
		WithArgs(args).
		Build(ctx)

	m, err := sdk.NewMachine(ctx, cfg, sdk.WithProcessRunner(cmd))
	if err != nil {
		return "", fmt.Errorf("failed to create new firecracker machine: %w", err)
	}

	err = m.Start(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create new firecracker machine: %w", err)
	}

	// Save the pid
	if err := f.ss.SavePID(cmd.Process.Pid); err != nil {
		return "", fmt.Errorf("saving pid %d to file: %w", cmd.Process.Pid, err)
	}

	// Save the config
	if err := f.ss.SaveVM(vm); err != nil {
		return "", fmt.Errorf("saving vm config to file: %w", err)
	}

	//START ----- for debugging debug
	// slog.Debug("FC instance info")
	// info, err := m.DescribeInstanceInfo(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("getting instance info: %w", err)
	// }
	// litter.Dump(info)
	// slog.Debug("FC machine cfg")
	// litter.Dump(m.Cfg)

	// meta := &metadata{}
	// if err := m.GetMetadata(ctx, meta); err != nil {
	// 	slog.Error("failed to get the metadata", "error", err.Error())
	// }
	// litter.Dump(meta)

	// END -------

	return "", nil
}
