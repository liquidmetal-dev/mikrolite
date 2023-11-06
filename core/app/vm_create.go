package app

import (
	"context"
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

	return nil, ErrNotImplemented
}
