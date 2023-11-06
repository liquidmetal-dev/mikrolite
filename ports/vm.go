package ports

import (
	"context"

	"github.com/mikrolite/mikrolite/domain"
)

// VMProvider represents a vmm implementation.
type VMrovider interface {
	// Create will create a new vm.
	Create(ctx context.Context, spec *domain.VMSpec) (string, error)
	// Stop will stop a running vm.
	Stop(ctx context.Context, id string) error
	// Delete will delete a running vm.
	Delete(ctx context.Context, id string) error
}
