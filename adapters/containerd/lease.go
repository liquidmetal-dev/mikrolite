package containerd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/namespaces"
)

func withNamespaceAndLease(ctx context.Context, owner string, manager leases.Manager) (context.Context, error) {
	nsCtx := namespaces.WithNamespace(ctx, Namespace)

	leaseName := leaseNameFromOwner(owner)

	lease, err := getLease(nsCtx, owner, manager)
	if err != nil {
		return ctx, fmt.Errorf("trying to get existing lease: %w", err)
	}

	if lease != nil {
		slog.Debug("Found existing lease", "name", leaseName)

		return leases.WithLease(nsCtx, lease.ID), nil
	}

	createdLease, err := createLease(nsCtx, owner, manager)
	if err != nil {
		return ctx, fmt.Errorf("creating new lease with name %s: %w", leaseName, err)
	}

	return leases.WithLease(nsCtx, createdLease.ID), nil
}

func getLease(ctx context.Context, owner string, manager leases.Manager) (*leases.Lease, error) {
	leaseName := leaseNameFromOwner(owner)

	filter := fmt.Sprintf("id==%s", leaseName)

	leases, err := manager.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing leases with filter %s: %w", filter, err)
	}

	for _, lease := range leases {
		if lease.ID == leaseName {
			return &lease, nil
		}
	}

	return nil, nil
}

func createLease(ctx context.Context, owner string, manager leases.Manager) (*leases.Lease, error) {
	leaseName := leaseNameFromOwner(owner)

	lease, err := manager.Create(ctx, leases.WithID(leaseName))
	if err != nil {
		return nil, fmt.Errorf("creating containerd lease with name %s: %w", leaseName, err)
	}

	return &lease, nil
}

func deleteLease(ctx context.Context, owner string, manager leases.Manager) error {
	lease, err := getLease(ctx, owner, manager)
	if err != nil {
		return nil
	}

	if err := manager.Delete(ctx, *lease, leases.SynchronousDelete); err != nil {
		return fmt.Errorf("deleting containerd lease %s: %w", lease.ID, err)
	}

	return nil
}

func leaseNameFromOwner(owner string) string {
	return fmt.Sprintf("mikrolite/%s", owner)
}
