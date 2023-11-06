package app

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

func (a *app) GetVM(ctx context.Context, name string) (*domain.VM, error) {
	return nil, ErrNotImplemented
}
