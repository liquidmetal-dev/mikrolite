package cloudinit

const (
	// InstanceDataKey is the metdata key name to use for instance data.
	InstanceDataKey = "meta-data"
	// UserdataKey is the metadata key name to use for user data.
	UserdataKey = "user-data"
	// VendorDataKey is the metadata key name to use for vendor data.
	VendorDataKey = "vendor-data"
	// NetworkConfigDataKey is the metadata key name for the network config.
	NetworkConfigDataKey = "network-config"
	// VolumeName is the name of a volume that contains cloud-init data.
	VolumeName = "CIDATA"
)

func IsCloudInitKey(keyName string) bool {
	switch keyName {
	case InstanceDataKey:
		return true
	case NetworkConfigDataKey:
		return true
	case UserdataKey:
		return true
	case VendorDataKey:
		return true
	default:
		return false
	}
}
