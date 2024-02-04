package app

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/spf13/afero"
)

// App represents the core application.
type App interface {
	ports.VMUseCases
}

func New(imageService ports.ImageService, vmService ports.VMProvider, stateService ports.StateService, fs afero.Fs, networkService ports.NetworkService) App {
	return &app{
		imageService:   imageService,
		fs:             fs,
		vmService:      vmService,
		stateService:   stateService,
		networkService: networkService,
	}
}

type app struct {
	imageService   ports.ImageService
	vmService      ports.VMProvider
	fs             afero.Fs
	stateService   ports.StateService
	networkService ports.NetworkService
}

type handlerInput struct {
	Owner             string
	SnapshotterKernel string
	SnapshotterVolume string
	VM                *domain.VM
}

type handler func(ctx context.Context, input *handlerInput) error
