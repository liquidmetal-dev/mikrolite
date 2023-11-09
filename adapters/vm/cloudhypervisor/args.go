package cloudhypervisor

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mikrolite/mikrolite/adapters/vm/shared"
	"github.com/mikrolite/mikrolite/core/domain"
)

func (p *provider) buildArgs(vm *domain.VM, cloudInitFile string) ([]string, error) {
	socketPath := p.socketPath()
	kernelPath := filepath.Join(vm.Status.KernelMount.Location, vm.Spec.Kernel.Source.Filename)

	args := []string{
		"--api-socket",
		socketPath,
		"--log-file",
		p.ss.LogPath(),
		"-v",
	}

	// Kernel and cmdline args
	if len(vm.Spec.Kernel.CmdLine) == 0 {
		vm.Spec.Kernel.CmdLine = defaultKernelCmdLine()
	}

	args = append(args, "--cmdline", shared.FormatKernelCmdLine(vm.Spec.Kernel.CmdLine))
	args = append(args, "--kernel", kernelPath)

	// CPU and memory
	args = append(args, "--cpus", fmt.Sprintf("boot=%d", vm.Spec.VCPU))
	args = append(args, "--memory", fmt.Sprintf("size=%dM", vm.Spec.MemoryInMb))

	// Volumes (root, additional, metadata)
	rootVolumeStatus, volumeStatusFound := vm.Status.VolumeMounts[vm.Spec.RootVolume.Name]
	if !volumeStatusFound {
		return nil, errors.New("root volume not found")
	}
	args = append(args, "--disk", fmt.Sprintf("path=%s", rootVolumeStatus.Location))
	args = append(args, fmt.Sprintf("path=%s,readonly=on", cloudInitFile))

	for id, vol := range vm.Status.VolumeMounts {
		if id == "root" {
			continue
		}
		args = append(args, fmt.Sprintf("path=%s", vol.Location))
	}

	// Network interfaces
	for name, _ := range vm.Spec.NetworkConfiguration.Interfaces {
		status, ok := vm.Status.NetworkStatus[name]
		if !ok {
			return nil, fmt.Errorf("failed to get network status for %s", name)
		}

		args = append(args, "--net")
		args = append(args, fmt.Sprintf("tap=%s,mac=%s", status.HostDeviveName, status.GuestMAC))

	}

	return args, nil

}

func (f *provider) socketPath() string {
	return filepath.Join(f.ss.Root(), "cloudhypervisor.sock")
}

func defaultKernelCmdLine() map[string]string {
	return map[string]string{
		"console": "hvc0",
		"root":    "/dev/vda",
		"rw":      "",
		"reboot":  "k",
		"panic":   "1",
		"ds":      "nocloud",
	}
}
