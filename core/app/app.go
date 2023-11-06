package app

import (
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/spf13/afero"
)

// App represents the core application.
type App interface {
	ports.VMUseCases
}

func New(imageService ports.ImageService, vmService ports.VMProvider, stateService ports.StateService, fs afero.Fs) App {
	return &app{
		imageService: imageService,
		fs:           fs,
		vmService:    vmService,
		stateService: stateService,
	}
}

type app struct {
	imageService ports.ImageService
	vmService    ports.VMProvider
	fs           afero.Fs
	stateService ports.StateService
}
