package app

import "github.com/mikrolite/mikrolite/core/ports"

// App represents the core application.
type App interface {
	ports.VMUseCases
}

func New(imageService ports.ImageService, vmService ports.VMProvider, fs ports.FileSystem) App {
	return &app{
		imageService: imageService,
		fs:           fs,
		vmService:    vmService,
	}
}

type app struct {
	imageService ports.ImageService
	fs           ports.FileSystem
	vmService    ports.VMProvider
}
