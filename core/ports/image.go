package ports

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

type PullAndMountInput struct {
	ImageName string
	Owner     string
	UsedFor   ImageUserFor
}

type ImageUserFor string

const (
	ImageUsedForVolume ImageUserFor = "volume"
	ImageUsedForKernel ImageUserFor = "kernel"
)

// Image service is the definition of a driven port for interacting with container images.
type ImageService interface {
	// PullAndMount will pull an image and mount it.
	PullAndMount(ctx context.Context, input PullAndMountInput) (*domain.Mount, error)
}
