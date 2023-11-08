package firecracker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	sdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sanity-io/litter"
	"github.com/spf13/afero"

	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/mikrolite/mikrolite/defaults"
)

const (
	ProviderName = "firecracker"
)

func New(stateService ports.StateService, fs afero.Fs) ports.VMProvider {
	return &Provider{
		ss: stateService,
		fs: fs,
	}
}

type Provider struct {
	ss ports.StateService
	fs afero.Fs
}

// Create will create a new vm.
func (f *Provider) Create(ctx context.Context, vm *domain.VM) (string, error) {
	socketPath := f.socketPath()
	kernelPath := filepath.Join(vm.Status.KernelMount.Location, vm.Spec.Kernel.Source.Filename)
	networkCfgPath := fmt.Sprintf("%s/fcnet.conflist", f.ss.Root())

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

	f.writeNetworkConfig(networkCfgPath, "fcnet")

	cfg := sdk.Config{
		VMID: vm.Name,
		//NetNS:           vm.Status.NetworkNamespace,
		SocketPath:      socketPath,
		KernelImagePath: kernelPath,
		KernelArgs:      formatKernelCmdLine(vm.Spec.Kernel.CmdLine),
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  intTo64Ptr(vm.Spec.VCPU),
			MemSizeMib: intTo64Ptr(vm.Spec.MemoryInMb),
			Smt:        boolPtr(true),
		},
		Drives:   []models.Drive{},
		LogPath:  f.ss.LogPath(), //TODO: should we change to use a fifo?
		LogLevel: "Debug",
	}

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
	netInt := sdk.NetworkInterface{
		StaticConfiguration: &sdk.StaticNetworkConfiguration{
			MacAddress:  vm.Status.NetworkStatus.GuestMAC,
			HostDevName: vm.Status.NetworkStatus.HostDeviveName,
		},
		AllowMMDS: true,
	}

	cfg.NetworkInterfaces = sdk.NetworkInterfaces{netInt}

	cfg.MmdsVersion = sdk.MMDSv1

	//TODO: this needs to be an optional arg for the path
	cmd := sdk.VMCommandBuilder{}.
		WithSocketPath(socketPath).
		WithBin("/home/richard/Downloads/firecracker-v1.5.0-x86_64/release-v1.5.0-x86_64/firecracker-v1.5.0-x86_64").
		WithStderr(stdErrFile).
		WithStdout(stdOutFile).
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
	slog.Debug("FC instance info")
	info, err := m.DescribeInstanceInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("getting instance info: %w", err)
	}
	litter.Dump(info)
	slog.Debug("FC machine cfg")
	litter.Dump(m.Cfg)

	// END -------

	return "", nil
}

// Stop will stop a running vm.
func (f *Provider) Stop(ctx context.Context, id string) error {
	return nil
}

// Delete will delete a running vm.
func (f *Provider) Delete(ctx context.Context, id string) error {
	return nil
}

func (f *Provider) ensureLogPath() error {
	logFile, err := f.fs.OpenFile(f.ss.LogPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaults.DataFilePerm)
	if err != nil {
		return fmt.Errorf("creating log file %s: %w", f.ss.LogPath(), err)
	}
	logFile.Close()

	return nil
}

func (f *Provider) writeNetworkConfig(path, networkName string) error {
	return os.WriteFile(path, []byte(fmt.Sprintf(`{
		"cniVersion": "0.3.1",
		"name": "%s",
		"plugins": [
		  {
			"type": "ptp",
			"ipam": {
			  "type": "host-local",
			  "subnet": "192.168.127.0/24"
			}
		  },
		  {
			"type": "tc-redirect-tap"
		  }
		]
	  }`, networkName)), 0644)
}

func (f *Provider) socketPath() string {
	return filepath.Join(f.ss.Root(), "firecracker.sock")
}

func formatKernelCmdLine(args map[string]string) string {
	output := []string{}

	for key, value := range args {
		if value == "" {
			output = append(output, key)
		} else {
			output = append(output, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return strings.Join(output, " ")
}
