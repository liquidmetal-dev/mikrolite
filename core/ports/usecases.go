package ports

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

// VMUseCases defines the uses cases related to interacting with vms
type VMUseCases interface {
	// CreateVM is the use case for creating a new VM.
	CreateVM(ctx context.Context, name string, owner string, vm *domain.VMSpec) (*domain.VM, error)
	// RemoveVM is the use case for removing a VM.
	RemoveVM(ctx context.Context, name string, owner string) error
	// GetVM is the use case for getting details of a VM.
	GetVM(ctx context.Context, name string) (*domain.VM, error)
	// ListVMs is the use case for listing all the vms.
	ListVMs(ctx context.Context) ([]*domain.VM, error)
}
