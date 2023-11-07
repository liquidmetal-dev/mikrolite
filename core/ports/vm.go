package ports

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

// VMProvider represents a vmm implementation.
type VMProvider interface {
	// Create will create a new vm.
	Create(ctx context.Context, vm *domain.VM) (string, error)
	// Stop will stop a running vm.
	Stop(ctx context.Context, id string) error
	// Delete will delete a running vm.
	Delete(ctx context.Context, id string) error
}
