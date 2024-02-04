package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/mikrolite/mikrolite/cloudinit"
	"github.com/mikrolite/mikrolite/core/domain"
	"github.com/mikrolite/mikrolite/core/ports"
	"github.com/mikrolite/mikrolite/defaults"
	"github.com/pterm/pterm"
	"github.com/spf13/afero"
	"github.com/yitsushi/macpot"
	"gopkg.in/yaml.v2"
)

func (a *app) CreateVM(ctx context.Context, input ports.CreateVMInput) (*domain.VM, error) {
	slog.Debug("Creating vm")

	if input.Name == "" {
		return nil, ErrNameRequired
	}

	if input.Spec == nil {
		return nil, ErrVmSpecRequired
	}

	//TODO: add validation

	vm, err := a.stateService.GetVM() //TODO: handle the state better
	if err != nil {
		return nil, fmt.Errorf("getting vm state: %w", err)
	}
	if vm != nil {
		return nil, ErrVMAlreadyExists
	}

	vm = &domain.VM{
		Name: input.Name,
	}
	vm.Spec = *input.Spec
	vm.Status = &domain.VMStatus{
		VolumeMounts: map[string]domain.Mount{},
	}
	vm.Status.NetworkNamespace = fmt.Sprintf("/var/run/netns/mikrolite-%s", input.Name)

	handlerInput := &handlerInput{
		SnapshotterVolume: input.SnapshotterVolume,
		SnapshotterKernel: input.SnapshotterKernal,
		Owner:             input.Owner,
		VM:                vm,
	}

	handlers := []handler{
		a.handleMetadataService,
		a.handleKernel,
		a.handleVolumes,
		a.handleNetwork,
		a.handleMetadata,
		a.handleVMCreateAndStart,
		a.handleFindIP,
		a.handleSaveVM,
	}

	for _, h := range handlers {
		if err := h(ctx, handlerInput); err != nil {
			return nil, err
		}
	}

	return vm, nil
}

func (a *app) handleVMCreateAndStart(ctx context.Context, input *handlerInput) error {
	_, err := a.vmService.Create(ctx, input.VM)
	if err != nil {
		return fmt.Errorf("creating vm: %w", err)
	}

	//TODO: add start if the provider supports start

	return nil
}

func (a *app) handleFindIP(ctx context.Context, input *handlerInput) error {
	mac := input.VM.Status.NetworkStatus["eth0"].GuestMAC

	sleep := 500 * time.Millisecond
	ip, err := retry[string](40, sleep, func() (string, error) {
		foundIp, foundErr := a.networkService.GetIPFromMac(mac)
		if foundErr != nil {
			return "", foundErr
		}
		if foundIp == "" {
			return "", errors.New("couldn't find ip address")
		}

		return foundIp, nil
	})
	if err != nil {
		return errors.New("failed to find ip address for vm")
	}

	input.VM.Status.IP = ip

	return nil
}

func (a *app) handleKernel(ctx context.Context, input *handlerInput) error {
	kernel := input.VM.Spec.Kernel
	if kernel.Source.HostPath != nil {
		input.VM.Status.KernelMount = &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: kernel.Source.HostPath.Path,
		}

		return nil
	}

	if kernel.Source.Container != nil {
		mount, err := a.imageService.PullAndMount(ctx, ports.PullAndMountInput{
			ImageName:   kernel.Source.Container.Image,
			Owner:       input.Owner,
			Snapshotter: input.SnapshotterKernel,
			ImageId:     "kernel",
		})
		if err != nil {
			return fmt.Errorf("getting kernel image: %w", err)
		}

		input.VM.Status.KernelMount = mount

		return nil
	}

	return errors.New("unexpected")
}

func (a *app) handleVolumes(ctx context.Context, input *handlerInput) error {
	pterm.DefaultSpinner.Info("ℹ️  Setting up volumes")

	slog.Debug("Setting up root volumes")
	rootVolumeMount, err := a.handleVolume(ctx, input.Owner, input.SnapshotterVolume, "root", &input.VM.Spec.RootVolume)
	if err != nil {
		return fmt.Errorf("handling root volume: %w", err)
	}
	input.VM.Status.VolumeMounts[input.VM.Spec.RootVolume.Name] = *rootVolumeMount

	slog.Debug("Setting up additional volumes")
	for i, vol := range input.VM.Spec.AdditionalVolumes {
		volumeId := fmt.Sprintf("vol%d", i)
		volMount, err := a.handleVolume(ctx, input.Owner, input.SnapshotterVolume, volumeId, &vol)
		if err != nil {
			return fmt.Errorf("handling volume %s: %s", vol.Name, err)
		}
		input.VM.Status.VolumeMounts[vol.Name] = *volMount
	}

	return nil
}

func (a *app) handleVolume(ctx context.Context, owner string, snapshotter string, volumeId string, volume *domain.Volume) (*domain.Mount, error) {
	if volume.Source.Raw != nil {
		return &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: volume.Source.Raw.Path,
		}, nil
	}

	if volume.Source.Container != nil {
		return a.imageService.PullAndMount(ctx, ports.PullAndMountInput{
			ImageName:   volume.Source.Container.Image,
			Owner:       owner,
			Snapshotter: snapshotter,
			ImageId:     volumeId,
		})
	}

	return nil, errors.New("unexpected")
}

func (a *app) handleMetadataService(ctx context.Context, input *handlerInput) error {
	if !a.vmService.HasMetadataService() {
		slog.Debug("vm provider doesn't have metadata service")

		return nil
	}

	metadataInt := &domain.NetwortInterface{
		GuestDeviceName:       "eth1",
		AllowMetadataRequests: true,
		AttachToBridge:        false,
		StaticIPv4Address: &domain.StaticIPv4Address{
			Address: "169.254.169.200/16",
			//Gateway: firecracker.String("169.254.169.254/16"),
		},
	}

	input.VM.Spec.NetworkConfiguration.Interfaces["eth1"] = *metadataInt

	return nil
}

func (a *app) handleNetwork(ctx context.Context, input *handlerInput) error {
	pterm.DefaultSpinner.Info("ℹ️  Setting up network")

	bridgeExists, err := a.networkService.BridgeExists(input.VM.Spec.NetworkConfiguration.BridgeName)
	if err != nil {
		return fmt.Errorf("checking if bridge exists")
	}
	// TODO: allow creation of bridge if it doesn't exists
	// if !bridgeExists {
	// 	if createErr := a.networkService.BridgeCreate(vm.Spec.NetworkConfig.BridgeName); createErr != nil {
	// 		return fmt.Errorf("creating bridge %s: %w", vm.Spec.NetworkConfig.BridgeName, err)
	// 	}
	// }
	if !bridgeExists {
		return errors.New("currently the network bridge must exist already. Create it using virt-manager/virsh")
	}

	input.VM.Status.NetworkStatus = map[string]domain.NetworkStatus{}
	for name, intCfg := range input.VM.Spec.NetworkConfiguration.Interfaces {
		slog.Debug("handling network interface", "name", name)

		mac, err := macpot.New(macpot.AsLocal(), macpot.AsUnicast())
		if err != nil {
			return fmt.Errorf("creating mac address vm: %w", err)
		}

		ifacePrefx := defaults.InterfacePrefix
		if intCfg.AllowMetadataRequests {
			ifacePrefx = defaults.MetadataInterfacePrefix
		}

		ifaceName, err := a.networkService.NewInterfaceName(ifacePrefx)
		if err != nil {
			return fmt.Errorf("getting vm network interface name: %s", err)
		}

		if createErr := a.networkService.InterfaceCreate(ifaceName, mac.ToString()); createErr != nil {
			return fmt.Errorf("creating vm network interface %s: %w", ifaceName, createErr)
		}

		if intCfg.AttachToBridge {
			if attachErr := a.networkService.AttachToBridge(ifaceName, input.VM.Spec.NetworkConfiguration.BridgeName); attachErr != nil {
				return fmt.Errorf("attching vm interface to bridge: %w", attachErr)
			}
		}

		input.VM.Status.NetworkStatus[name] = domain.NetworkStatus{
			HostDeviveName: ifaceName,
			GuestMAC:       mac.ToString(),
		}
	}

	return nil
}

func (a *app) handleMetadata(ctx context.Context, input *handlerInput) error {
	networkConfig, err := generateNetworkConfig(input.VM)
	if err != nil {
		return fmt.Errorf("generating network config")
	}
	input.VM.Status.Metadata = map[string]string{
		cloudinit.NetworkConfigDataKey: networkConfig,
	}

	metadata, err := a.createMetadata(input.VM)
	if err != nil {
		return fmt.Errorf("generating metada data: %w", err)
	}
	input.VM.Status.Metadata[cloudinit.InstanceDataKey] = metadata

	if input.VM.Spec.Bootstrap != nil {

		userdata, err := a.createUserData(input.VM)
		if err != nil {
			return fmt.Errorf("generating user data: %w", err)
		}

		input.VM.Status.Metadata[cloudinit.UserdataKey] = userdata
	}

	return nil

}

func (a *app) handleSaveVM(ctx context.Context, input *handlerInput) error {
	return a.stateService.SaveVM(input.VM)
}

func (a *app) createMetadata(vm *domain.VM) (string, error) {
	metadata := map[string]string{}
	metadata["instance_id"] = vm.Name
	metadata["cloud_name"] = "mikrolite"

	data, err := yaml.Marshal(&metadata)
	if err != nil {
		return "", fmt.Errorf("marshalling metadata: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (a *app) createUserData(vm *domain.VM) (string, error) {
	userdata := &cloudinit.UserData{
		FinalMessage: "mikrolite booted system",
		BootCommands: []string{
			"ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf",
		},
		HostName: vm.Name,
	}

	if vm.Spec.Bootstrap.SSHKey != "" {
		data, err := afero.ReadFile(a.fs, vm.Spec.Bootstrap.SSHKey)
		if err != nil {
			return "", fmt.Errorf("reading ssh key %s: %w", vm.Spec.Bootstrap.SSHKey, err)
		}
		user := cloudinit.User{
			Name:              "ml",
			Gecos:             "Mikrolite user",
			Shell:             "/bin/bash",
			Groups:            "sudo",
			Sudo:              "ALL=(ALL) NOPASSWD:ALL",
			SSHAuthorizedKeys: []string{string(data)},
		}

		userdata.Users = []cloudinit.User{user}
	}

	data, err := yaml.Marshal(userdata)
	if err != nil {
		return "", fmt.Errorf("marshalling userdate: %w", err)
	}
	dataWithHeader := append([]byte("## template: jinja\n#cloud-config\n\n"), data...)

	return base64.StdEncoding.EncodeToString(dataWithHeader), nil

}

func generateNetworkConfig(vm *domain.VM) (string, error) {
	netConf := &cloudinit.Network{
		Version:  2,
		Ethernet: map[string]cloudinit.Ethernet{},
	}

	for name, netInt := range vm.Spec.NetworkConfiguration.Interfaces {
		status, ok := vm.Status.NetworkStatus[name]
		if !ok {
			return "", fmt.Errorf("failed to get network status for %s", name)
		}

		eth := &cloudinit.Ethernet{
			Match: cloudinit.Match{
				MACAddress: status.GuestMAC,
			},
			DHCP4:          firecracker.Bool(true),
			DHCPIdentifier: firecracker.String(cloudinit.DhcpIdentifierMac),
		}

		if netInt.StaticIPv4Address != nil {
			if err := addStaticIP(netInt.StaticIPv4Address, eth); err != nil {
				return "", fmt.Errorf("adding static ipv4 config: %w", err)
			}
		}

		netConf.Ethernet[netInt.GuestDeviceName] = *eth
	}

	nd, err := yaml.Marshal(netConf)
	if err != nil {
		return "", fmt.Errorf("marshalling network data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(nd), nil
}

func addStaticIP(ipConfig *domain.StaticIPv4Address, eth *cloudinit.Ethernet) error {
	eth.DHCP4 = firecracker.Bool(false)
	eth.Addresses = []string{ipConfig.Address}

	if ipConfig.Gateway != nil && *ipConfig.Gateway != "" {
		gwIp, err := getIPFromCIDR(*ipConfig.Gateway)
		if err != nil {
			return fmt.Errorf("failed to get IP from cidr %s: %w", *ipConfig.Gateway, err)
		}

		eth.GatewayIPv4 = gwIp
		// eth.Routes = []cloudinit.Routes{
		// 	{
		// 		To:     "default",
		// 		Via:    gwIp,
		// 		OnLink: firecracker.Bool(true),
		// 		Metric: firecracker.Int(100),
		// 	},
		// }
	}

	if len(ipConfig.Nameservers) == 0 {
		return nil
	}

	eth.Nameservers = cloudinit.Nameservers{
		Addresses: eth.Nameservers.Addresses,
	}

	return nil
}

func getIPFromCIDR(cidr string) (string, error) {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		return "", fmt.Errorf("parsing cidr: %w", err)
	}

	slashIndex := strings.Index(cidr, "/")

	return cidr[:slashIndex], nil
}

// TODO: change this to timeout instead of number of reties
func retry[T any](attempts int, sleep time.Duration, f func() (T, error)) (result T, err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			//sleepInSeconds *= 2
		}
		result, err = f()
		if err == nil {
			return result, nil
		}
	}
	return result, fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
