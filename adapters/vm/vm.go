package vm

import (
	"github.com/mikrolite/mikrolite/adapters/vm/firecracker"
	"github.com/mikrolite/mikrolite/ports"
)

func New(name string) (ports.VMProvider, error) {
	switch name {
	case firecracker.ProviderName:
		return &firecracker.Provider{}, nil
	default:
		return nil, NewUnknownProvider(name)
	}
}
