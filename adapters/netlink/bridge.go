package netlink

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/vishvananda/netlink"
)

func (s *networkService) BridgeCreate(name string) error {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Creating network bridge: %s\n", name))

	la := netlink.NewLinkAttrs()
	la.Name = name
	bridge := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(bridge); err != nil {
		return fmt.Errorf("creating bridge %s: %w", name, err)
	}

	return nil
}

func (s *networkService) BridgeDelete(name string) error {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Deleting network bridge: %s\n", name))

	return deleteLink(name, "bridge")
}

func (s *networkService) BridgeExists(name string) (bool, error) {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Checking if bridge exists: %s\n", name))

	return linkExists(name, "bridge")
}

func (s *networkService) AttachToBridge(interfaceName string, bridgeName string) error {
	pterm.DefaultSpinner.Info(fmt.Sprintf("ℹ️  Adding network interface \"%s\" to bridge \"%s\"\n", interfaceName, bridgeName))

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
