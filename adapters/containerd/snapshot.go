package containerd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/opencontainers/image-spec/identity"

	"github.com/mikrolite/mikrolite/core/domain"
)

func (s *imageService) snapshotImage(ctx context.Context, owner string, image containerd.Image, snapshotter string, imageId string) (*domain.Mount, error) {
	content, err := image.RootFS(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting root filesystem for image %s: %w", image.Name(), err)
	}

	parent := identity.ChainID(content).String()

	snapshotKey := fmt.Sprintf("mikrolite/%s/%s", owner, imageId)
	snapshotClient := s.client.SnapshotService(snapshotter)

	exists, err := snapshotExists(ctx, snapshotKey, snapshotClient)
	if err != nil {
		return nil, fmt.Errorf("checking if snapshot %s exists: %w", snapshotKey, err)
	}

	mounts := []mount.Mount{}
	if exists {
		mounts, err = snapshotClient.Mounts(ctx, snapshotKey)
		if err != nil {
			return nil, fmt.Errorf("getting mounts for snapshot %s: %w", snapshotKey, err)
		}
	} else {
		mounts, err = snapshotClient.Prepare(ctx, snapshotKey, parent)
		if err != nil {
			return nil, fmt.Errorf("preparing mounts for snapshot %s: %w", snapshotKey, err)
		}
	}

	switch snapshotter {
	case "native":
		return &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: mounts[0].Source,
		}, nil
	case "devmapper":
		return &domain.Mount{
			Type:     domain.MountTypeBlockDevice,
			Location: mounts[0].Source,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported snapshotter %s encountered", snapshotter)
	}
}

func snapshotExists(ctx context.Context, key string, snapshotter snapshots.Snapshotter) (bool, error) {
	exists := false
	err := snapshotter.Walk(ctx, func(ctx context.Context, i snapshots.Info) error {
		if i.Name == key {
			exists = true
			return nil
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("listing snapshots: %w", err)
	}

	return exists, nil
}

func ensureUnpacked(ctx context.Context, image containerd.Image, snapshotter string) error {
	isUnpacked, err := image.IsUnpacked(ctx, snapshotter)
	if err != nil {
		return fmt.Errorf("checking if image %s is unpacked with snapshotter %s: %w", image.Name(), snapshotter, err)
	}

	if isUnpacked {
		slog.Debug("image is already unpacked", "image", image.Name(), "snapshotter", snapshotter)

		return nil
	}

	slog.Debug("image isn't already unpacked, unpacking", "image", image.Name(), "snapshotter", snapshotter)
	if err := image.Unpack(ctx, snapshotter); err != nil {
		return fmt.Errorf("unpacking image %s with snapshotter %s: %w", image.Name(), snapshotter, err)
	}

	return nil
}
