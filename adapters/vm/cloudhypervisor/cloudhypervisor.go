package cloudhypervisor

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/mikrolite/mikrolite/adapters/vm/shared"
	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/mikrolite/mikrolite/defaults"
	"github.com/spf13/afero"
)

const (
	ProviderName = "cloudhypervisor"
)

func New(binaryPath string, stateService ports.StateService, ds ports.DiskService, fs afero.Fs) ports.VMProvider {
	return &provider{
		ss:         stateService,
		fs:         fs,
		ds:         ds,
		binaryPath: binaryPath,
	}
}

type provider struct {
	ss         ports.StateService
	ds         ports.DiskService
	fs         afero.Fs
	binaryPath string
}

func (f *provider) Create(ctx context.Context, vm *domain.VM) (string, error) {

	cloudInitFile, err := shared.CreateCloudInitImage(ctx, true, vm, f.ss, f.ds)
	if err != nil {
		return "", fmt.Errorf("creating cloud-init disk image: %w", err)
	}

	args, err := f.buildArgs(vm, cloudInitFile)
	if err != nil {
		return "", fmt.Errorf("building cloud hypervisor args: %w", err)
	}

	cmd := exec.Command(f.binaryPath, args...)

	stdOutFile, err := f.fs.OpenFile(f.ss.StdoutPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaults.DataFilePerm)
	if err != nil {
		return "", fmt.Errorf("opening stdout file %s: %w", f.ss.StdoutPath(), err)
	}

	stdErrFile, err := f.fs.OpenFile(f.ss.StderrPath(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, defaults.DataFilePerm)
	if err != nil {
		return "", fmt.Errorf("opening sterr file %s: %w", f.ss.StderrPath(), err)
	}

	cmd.Stderr = stdErrFile
	cmd.Stdout = stdOutFile
	cmd.Stdin = &bytes.Buffer{}

	if startErr := cmd.Start(); startErr != nil {
		return "", fmt.Errorf("starting cloudhypervisor: %w", err)
	}

	// Save the pid
	if err := f.ss.SavePID(cmd.Process.Pid); err != nil {
		return "", fmt.Errorf("saving pid %d to file: %w", cmd.Process.Pid, err)
	}

	// Save the config
	if err := f.ss.SaveVM(vm); err != nil {
		return "", fmt.Errorf("saving vm config to file: %w", err)
	}

	return "", nil
}

func (f *provider) Stop(ctx context.Context, name string) error {
	//TODO: call the api to stop CH

	pid, err := f.ss.GetPID()
	if err != nil {
		return fmt.Errorf("getting vm pid: %w", err)
	}

	if pid == 0 {
		slog.Debug("pid not set for vm, skipping stop", "name", name)

		return nil
	}

	p, _ := os.FindProcess(pid)
	p.Signal(syscall.SIGHUP)

	return nil
}

func (f *provider) Delete(ctx context.Context, name string) error {
	pid, err := f.ss.GetPID()
	if err != nil {
		return fmt.Errorf("getting vm pid: %w", err)
	}

	if pid == 0 {
		slog.Debug("pid not set for vm, skipping stop", "name", name)

		return nil
	}

	if err := shared.StopProcess(ctx, pid); err != nil {
		return fmt.Errorf("stopping firecracker process: %w", err)
	}

	return nil
}

func (f *provider) HasMetadataService() bool {
	return false
}
