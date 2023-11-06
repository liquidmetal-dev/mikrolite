package app

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

func (a *app) CreateVM(ctx context.Context, vm *domain.VM) (*domain.VM, error) {
	return nil, ErrNotImplemented
}
