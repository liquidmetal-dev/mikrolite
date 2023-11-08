package domain

// VM represents the spec and status of a VM.
type VM struct {
	// Name is the name of the vm. Used as an identified only and not the hostname.
	Name string `json:"name"`

	// Spec is the specification of the vm
	Spec VMSpec `json:"spec"`

	// Status holds runtime status information for the vm
	Status *VMStatus `json:"status,omitempty"`
}

// VMSpec represents the specification of a VM.
type VMSpec struct {
	// Kernel defines the kernel to use.
	Kernel Kernel `json:"kernel"`
	// RootVolume defines the root volume.
	RootVolume Volume `json:"root_volume"`
	// AdditionalVolumes defines any other volumes.
	AdditionalVolumes []Volume `json:"additional_volumes"`
	// VCPU defines how many virtual cpus the vm should have.
	VCPU int `json:"vcpu"`
	// MemoryInMb defines how much memory the vm should have.
	MemoryInMb int `json:"memory_in_mb"`
	// NetworkConfig holds the configuration for the the vm networking
	NetworkConfig NetworkConfiguration `json:"network_config"`
}

// VMStatus holds the runtime status information of the vm.
type VMStatus struct {
	// VolumeMounts holds details of where the volumes are mounted.
	VolumeMounts map[string]Mount `json:"volume_mounts"`

	// KernelMount holds the mount details for the kernel.
	KernelMount *Mount `json:"kernel_mount,omitempty"`

	// NetworkNamespace is the netns for this vm
	NetworkNamespace string

	// NetworkStatus hols the status of the network.
	NetworkStatus *NetworkStatus `json:"network_status,omitempty"`
}

// Kernel defines the kernel to use.
type Kernel struct {
	// Source defines where to get the kernel from.
	Source KernelSource `json:"source"`
	// CmdLine is the cmd line args for the kernel.
	CmdLine map[string]string `json:"cmd_line"`
}

// Volume represents a volume for a VM.
type Volume struct {
	// Name is the name of the volume, used as an identifier only.
	Name string `json:"name"`
	// Source specifies the source of the volume.
	Source VolumeSource `json:"source"`
}

// VolumeSource is the source of a volumes.
type VolumeSource struct {
	// Container is used to specify the volume comes from a container.
	Container *ContainerVolumeSource `json:"container,omitempty"`
	// Raw is used to specify a volume comes from a raw fs file.
	Raw *RawVolumeSource `json:"raw,omitempty"`
}

// ContainerVolumeSource is the specification of using a container as a volume.
type ContainerVolumeSource struct {
	// Image is the container image name.
	Image string `json:"image"`
}

// RawVolumeSource is the specification of using a raw file for the source of a volume.
type RawVolumeSource struct {
	// Path is the path to the raw fs file.
	Path string `json:"path"`
}

// KernelSource is the source of the kernel.
type KernelSource struct {
	// Filename is the name of the kernel file in the image or in the hostpath
	Filename string `json:"filename"`
	// Container specifies the kernel comes from a container image.
	Container *ContainerKernelSource `json:"container,omitempty"`
	// HostPath specified the kernel comes from a file already on the host system
	HostPath *HostPathKernelSource `json:"host_path,omitempty"`
}

// ContainerKernelSource holds the speciofication of using a container for a kernel.
type ContainerKernelSource struct {
	// Image is the container images that holds the kernel.
	Image string `json:"image"`
}

// HostPathKernelSource is teh specification of using a file on the host for the kernel.
type HostPathKernelSource struct {
	// Path is the path of the host to the kernel to use.
	Path string `json:"path"`
}

// NetworkConfiguration is network configuration for the VM.
type NetworkConfiguration struct {
	// BridgeName is the name of the bridge to attach the vm to
	BridgeName string `json:"bridge_name"`
}

// NetworkStatus holds information about the status of the network
type NetworkStatus struct {
	// GuestMAC is the mac address to use.
	GuestMAC string `json:"guest_mac"`
	// GuestDeviceName is the name of the device on the guest.
	GuestDeviceName string `json:"guest_device_name"`
	// HostDeviceName is the name of the network device on the host
	HostDeviveName string `json:"host_device_name"`
}

// Mount containes details of a mount.
type Mount struct {
	// Type is the type of the mount.
	Type MountType `json:"type"`
	// Location is the location of the mount.
	Location string `json:"location"`
}

// MountType is the type of volume mount.
type MountType string

const (
	// MountTypeBlockDevice is a block device mount.
	MountTypeBlockDevice MountType = "blockdevice"
	// MountTypeFilesystemPath is a filesystem mount.
	MountTypeFilesystemPath MountType = "filesystem"
)
