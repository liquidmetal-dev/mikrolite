package vm

import (
	"github.com/mikrolite/mikrolite/adapters/vm/firecracker"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/spf13/afero"
)

func New(name string, stateService ports.StateService, fs afero.Fs) (ports.VMProvider, error) {
	switch name {
	case firecracker.ProviderName:
		return firecracker.New(stateService, fs), nil
	default:
		return nil, NewUnknownProvider(name)
	}
}
