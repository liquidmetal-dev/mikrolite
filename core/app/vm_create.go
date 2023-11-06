package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
)

func (a *app) CreateVM(ctx context.Context, name string, owner string, vm *domain.VMSpec) (*domain.VM, error) {
	slog.Debug("Creating vm")

	if name == "" {
		return nil, ErrNameRequired
	}

	if vm == nil {
		return nil, ErrVmSpecRequired
	}

	//TODO: add validation

	storedVM, err := a.stateService.GetVM()
	if err != nil {
		return nil, fmt.Errorf("getting vm state: %w", err)
	}
	if storedVM != nil {
		return nil, fmt.Errorf("vm %s already exists", name)
	}

	kernelMount, err := a.handleKernel(ctx, owner, &vm.Kernel)
	if err != nil {
		return nil, fmt.Errorf("handling kernel: %w", err)
	}
	fmt.Println(kernelMount)

	rootVolumeMount, err := a.handleVolume(ctx, owner, &vm.RootVolume)
	if err != nil {
		return nil, fmt.Errorf("handling root volume: %w", err)
	}
	fmt.Println(rootVolumeMount)

	return nil, nil
}

func (a *app) handleKernel(ctx context.Context, owner string, kernel *domain.Kernel) (*domain.Mount, error) {
	if kernel.Source.HostPath != nil {
		return &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: kernel.Source.HostPath.Path,
		}, nil
	}

	if kernel.Source.Container != nil {
		return a.imageService.PullAndMount(ctx, ports.PullAndMountInput{
			ImageName: kernel.Source.Container.Image,
			Owner:     owner,
			UsedFor:   ports.ImageUsedForKernel,
		})
	}

	return nil, errors.New("unexpected")
}

func (a *app) handleVolume(ctx context.Context, owner string, volume *domain.Volume) (*domain.Mount, error) {
	if volume.Source.Raw != nil {
		return &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: volume.Source.Raw.Path,
		}, nil
	}

	if volume.Source.Container != nil {
		return a.imageService.PullAndMount(ctx, ports.PullAndMountInput{
			ImageName: volume.Source.Container.Image,
			Owner:     owner,
			UsedFor:   ports.ImageUsedForVolume,
		})
	}

	return nil, errors.New("unexpected")
}
