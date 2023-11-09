package defaults

const (
	// DataFilePerm is the permissions to use for data files.
	DataFilePerm = 0o644

	// SharedBridgeName is the name of the bridge to use when none is specified.
	SharedBridgeName = "mikrolite"

	// InterfacePrefix is a prefix to use for network interface names
	InterfacePrefix = "mlt"

	// MetadataInterfacePrefix is a prefix to use for network interface names for a metadata connection
	MetadataInterfacePrefix = "mltm"
)
