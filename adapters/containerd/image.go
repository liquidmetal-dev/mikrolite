package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/namespaces"
	"github.com/pterm/pterm"

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
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Pulling and mounting image: %s\n", input.ImageName))

	leaseCtx, err := withNamespaceAndLease(ctx, input.Owner, s.client.LeasesService())
	if err != nil {
		return nil, fmt.Errorf("setting namespace and lease: %w", err)
	}

	image, err := s.client.Pull(leaseCtx, input.ImageName)
	if err != nil {
		return nil, fmt.Errorf("pulling image %s: %w", input.ImageName, err)
	}

	if err := ensureUnpacked(leaseCtx, image, input.Snapshotter); err != nil {
		return nil, fmt.Errorf("ensuring image is unpacked: %w", err)
	}

	mount, err := s.snapshotImage(leaseCtx, input.Owner, image, input.Snapshotter, input.ImageId)
	if err != nil {
		return nil, fmt.Errorf("snapshotting image %s: %w", image.Name(), err)
	}

	return mount, nil
}

func (s *imageService) Cleanup(ctx context.Context, owner string) error {
	pterm.DefaultSpinner.Info("ℹ️  Cleaning up images")

	nsCtx := namespaces.WithNamespace(ctx, Namespace)
	leaseName := leaseNameFromOwner(owner)
	lease := leases.Lease{ID: leaseName}

	if err := s.client.LeasesService().Delete(nsCtx, lease, leases.SynchronousDelete); err != nil {
		return fmt.Errorf("deleting containerd lease %s: %w", leaseName, err)
	}

	return nil
}
