package netlink

import (
	"fmt"
	"log/slog"

	"github.com/vishvananda/netlink"
)

func (s *networkService) BridgeCreate(name string) error {
	slog.Info("Creating network bridge", "name", name)

	la := netlink.NewLinkAttrs()
	la.Name = name
	bridge := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("creating bridge %s: %w", name, err)
	}

	return nil
}

func (s *networkService) BridgeDelete(name string) error {
	slog.Info("Deleting network bridge", "name", name)

	return deleteLink(name, "bridge")
}

func (s *networkService) BridgeExists(name string) (bool, error) {
	slog.Info("Checking if bridge exists", "name", name)

	return linkExists(name, "bridge")
}

func (s *networkService) AttachToBridge(interfaceName string, bridgeName string) error {
	slog.Info("Adding network interface to bridge", "interface", interfaceName, "bridge", bridgeName)

	bridgeLink, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("getting bridge %s: %w", bridgeName, err)
	}

	interfaceLink, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("getting interface %s: %w", interfaceName, err)
	}

	if err := netlink.LinkSetMaster(interfaceLink, bridgeLink); err != nil {
		return fmt.Errorf("adding interface to bridge: %w", err)
	}

	return nil
}
