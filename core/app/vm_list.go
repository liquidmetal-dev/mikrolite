package app

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

func (a *app) ListVMs(ctx context.Context) ([]*domain.VM, error) {
	return nil, ErrNotImplemented
}
