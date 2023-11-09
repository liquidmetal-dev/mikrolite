package vm

import (
	"github.com/mikrolite/mikrolite/adapters/vm/firecracker"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/spf13/afero"
)

func New(name string, stateService ports.StateService, diskSvc ports.DiskService, fs afero.Fs) (ports.VMProvider, error) {
	switch name {
	case firecracker.ProviderName:
		return firecracker.New(stateService, diskSvc, fs), nil
	default:
		return nil, NewUnknownProvider(name)
	}
}
