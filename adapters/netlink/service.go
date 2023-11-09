package netlink

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/vishvananda/netlink"
)

func New() ports.NetworkService {
	return &networkService{}
}

type networkService struct {
}

func deleteLink(name string, linkType string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		if errors.As(err, &netlink.LinkNotFoundError{}) {
			slog.Debug("network link not found, skipping deletion", "name", name, "type", linkType)

			return nil
		}

		return fmt.Errorf("looking up %s %s: %w", linkType, name, err)
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("deleting %s %s: %w", linkType, name, err)
	}

	return nil
}

func linkExists(name string, linkType string) (bool, error) {
	if _, err := netlink.LinkByName(name); err != nil {
		if errors.As(err, &netlink.LinkNotFoundError{}) {
			slog.Debug("network link doesn't exist", "name", name, "type", linkType)

			return false, nil
		}

		return false, fmt.Errorf("looking up %s: %w", name, err)
	}

	slog.Debug("network link exists", "name", name, "type", linkType)
	return true, nil
}
