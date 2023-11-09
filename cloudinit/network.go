package cloudinit

const (
	DhcpIdentifierMac = "mac"
)

type Network struct {
	Version  int                 `yaml:"version"`
	Ethernet map[string]Ethernet `yaml:"ethernets"`
}

type Ethernet struct {
	Match          Match       `yaml:"match"`
	Addresses      []string    `yaml:"addresses,omitempty"`
	GatewayIPv4    string      `yaml:"gateway4,omitempty"`
	DHCP4          *bool       `yaml:"dhcp4,omitempty"`
	DHCPIdentifier *string     `yaml:"dhcp-identifier,omitempty"`
	Nameservers    Nameservers `yaml:"nameservers,omitempty"`
	Routes         []Routes    `yaml:"routes,omitempty"`
}

type Match struct {
	MACAddress string `yaml:"macaddress,omitempty"`
	Name       string `yaml:"name,omitempty"`
}

type Nameservers struct {
	Search    []string `yaml:"search,omitempty"`
	Addresses []string `yaml:"addresses,omitempty"`
}

type Routes struct {
	To     string `yaml:"to"`
	Via    string `yaml:"via"`
	Metric *int   `yaml:"metric,omitempty"`
	OnLink *bool  `yaml:"on-link,omitempty"`
}
