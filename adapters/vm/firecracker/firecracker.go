package firecracker

import (
	"context"

	"github.com/mikrolite/mikrolite/domain"
)

const (
	ProviderName = "firecracker"
)

type Provider struct {
}

// Create will create a new vm.
func (f *Provider) Create(ctx context.Context, spec *domain.VMSpec) (string, error) {
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
