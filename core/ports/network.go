package ports

type NetworkService interface {
	BridgeCreate(name string) error
	BridgeDelete(name string) error
	BridgeExists(name string) (bool, error)

	InterfaceCreate(name string, mac string) error
	InterfaceDelete(name string) error
	InterfaceExists(name string) (bool, error)

	AttachToBridge(interfaceName string, bridgeName string) error

	NewInterfaceName(prefix string) (string, error)

	GetIPFromMac(macAddress string) (string, error)
}
