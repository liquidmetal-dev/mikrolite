package app

import "github.com/mikrolite/mikrolite/core/ports"

// App represents the core application.
type App interface {
	ports.VMUseCases
}

func New() App {
	return &app{}
}

type app struct {
}
