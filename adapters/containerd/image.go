package containerd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/containerd/containerd"

	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
)

func NewImageService(client *containerd.Client) ports.ImageService {
	return &imageService{
		client: client,
	}
}

type imageService struct {
	client *containerd.Client
}

func (s *imageService) PullAndMount(ctx context.Context, input ports.PullAndMountInput) (*domain.Mount, error) {
	slog.Info("Pulling and mounting image", "image", input.ImageName)

	leaseCtx, err := withNamespaceAndLease(ctx, input.Owner, s.client.LeasesService())
	if err != nil {
		return nil, fmt.Errorf("setting namespace and lease: %w", err)
	}

	image, err := s.client.Pull(leaseCtx, input.ImageName)
	if err != nil {
		return nil, fmt.Errorf("pulling image %s: %w", input.ImageName, err)
	}

	if err := ensureUnpacked(leaseCtx, image, input.UsedFor); err != nil {
		return nil, fmt.Errorf("ensuring image is unpacked: %w", err)
	}

	mount, err := s.snapshotImage(leaseCtx, input.Owner, image, input.UsedFor)
	if err != nil {
		return nil, fmt.Errorf("snapshotting image %s: %w", image.Name(), err)
	}

	return mount, nil
}
