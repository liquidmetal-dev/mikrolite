package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mikrolite/mikrolite/core/domain"
)

func (a *app) CreateVM(ctx context.Context, name string, vm *domain.VMSpec) (*domain.VM, error) {
	slog.Debug("Creating vm")

	if name == "" {
		return nil, ErrNameRequired
	}

	if vm == nil {
		return nil, ErrVmSpecRequired
	}

	//TODO: add validation

	//TODO: check if vm already exists
	storedVM, err := a.stateService.GetVM()
	if err != nil {
		return nil, fmt.Errorf("getting vm state: %w", err)
	}
	if storedVM != nil {
		return nil, fmt.Errorf("vm %s already exists", name)
	}

	return nil, ErrNotImplemented
}
