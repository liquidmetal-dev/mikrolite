package vm

import (
	"errors"

	"github.com/mikrolite/mikrolite/adapters/vm/cloudhypervisor"
	"github.com/mikrolite/mikrolite/adapters/vm/firecracker"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/spf13/afero"
)

type VMProviderProps struct {
	StateService       ports.StateService
	DiskSvc            ports.DiskService
	Fs                 afero.Fs
	FirecrackerBin     string
	CloudHypervisorBin string
}

func New(name string, props VMProviderProps) (ports.VMProvider, error) {
	switch name {
	case firecracker.ProviderName:
		if props.FirecrackerBin == "" {
			return nil, errors.New("must supply a path to a firecracker binary")
		}

		return firecracker.New(props.FirecrackerBin, props.StateService, props.DiskSvc, props.Fs), nil
	case cloudhypervisor.ProviderName:
		if props.CloudHypervisorBin == "" {
			return nil, errors.New("must supply a path to a cloud hypervisor binary")
		}

		return cloudhypervisor.New(props.CloudHypervisorBin, props.StateService, props.DiskSvc, props.Fs), nil
	default:
		return nil, NewUnknownProvider(name)
	}
}
