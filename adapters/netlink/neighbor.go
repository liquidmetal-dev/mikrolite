package netlink

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func (s *networkService) GetIPFromMac(macAddress string) (string, error) {
	toFind, err := net.ParseMAC(macAddress)
	if err != nil {
		return "", fmt.Errorf("parsing mac address: %s", err)
	}

	neighbors, err := netlink.NeighList(0, netlink.FAMILY_V4)
	if err != nil {
		return "", fmt.Errorf("getting ip neighbors: %w", err)
	}

	for _, n := range neighbors {
		if n.HardwareAddr.String() == toFind.String() {
			return n.IP.String(), nil
		}
	}

	return "", nil
}
