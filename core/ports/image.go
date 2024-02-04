package ports

import (
	"context"

	"github.com/mikrolite/mikrolite/core/domain"
)

type PullAndMountInput struct {
	ImageId     string
	ImageName   string
	Owner       string
	Snapshotter string
}

// Image service is the definition of a driven port for interacting with container images.
type ImageService interface {
	// PullAndMount will pull an image and mount it.
	PullAndMount(ctx context.Context, input PullAndMountInput) (*domain.Mount, error)

	// Cleanup any images used by a VM.
	Cleanup(ctx context.Context, owner string) error
}
