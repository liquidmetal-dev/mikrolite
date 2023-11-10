package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

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

	handlers := []handler{
		a.handleMetadataService,
		a.handleKernel,
		a.handleVolumes,
		a.handleNetwork,
		a.handleMetadata,
		a.handleVMCreateAndStart,
	}

	for _, h := range handlers {
		if err := h(ctx, input.Owner, vm); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (a *app) handleVMCreateAndStart(ctx context.Context, owner string, vm *domain.VM) error {
	_, err := a.vmService.Create(ctx, vm)
	if err != nil {
		return fmt.Errorf("creating vm: %w", err)
	}

	//TODO: add start if the provider supports start

	return nil
}

func (a *app) handleKernel(ctx context.Context, owner string, vm *domain.VM) error {
	kernel := vm.Spec.Kernel
	if kernel.Source.HostPath != nil {
		vm.Status.KernelMount = &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: kernel.Source.HostPath.Path,
		}

		return nil
	}

	if kernel.Source.Container != nil {
		mount, err := a.imageService.PullAndMount(ctx, ports.PullAndMountInput{
			ImageName: kernel.Source.Container.Image,
			Owner:     owner,
			UsedFor:   ports.ImageUsedForKernel,
		})
		if err != nil {
			return fmt.Errorf("getting kernel image: %w", err)
		}

		vm.Status.KernelMount = mount

		return nil
	}

	return errors.New("unexpected")
}

func (a *app) handleVolumes(ctx context.Context, owner string, vm *domain.VM) error {
	pterm.DefaultSpinner.Info("ℹ️  Setting up volumes")

	slog.Debug("Setting up root volumes")
	rootVolumeMount, err := a.handleVolume(ctx, owner, &vm.Spec.RootVolume)
	if err != nil {
		return fmt.Errorf("handling root volume: %w", err)
	}
	vm.Status.VolumeMounts[vm.Spec.RootVolume.Name] = *rootVolumeMount

	slog.Debug("Setting up additional volumes")
	for _, vol := range vm.Spec.AdditionalVolumes {
		volMount, err := a.handleVolume(ctx, owner, &vol)
		if err != nil {
			return fmt.Errorf("handling volume %s: %s", vol.Name, err)
		}
		vm.Status.VolumeMounts[vol.Name] = *volMount
	}

	return nil
}

func (a *app) handleVolume(ctx context.Context, owner string, volume *domain.Volume) (*domain.Mount, error) {
	if volume.Source.Raw != nil {
		return &domain.Mount{
			Type:     domain.MountTypeFilesystemPath,
			Location: volume.Source.Raw.Path,
		}, nil
	}

	if volume.Source.Container != nil {
		return a.imageService.PullAndMount(ctx, ports.PullAndMountInput{
			ImageName: volume.Source.Container.Image,
			Owner:     owner,
			UsedFor:   ports.ImageUsedForVolume,
		})
	}

	return nil, errors.New("unexpected")
}

func (a *app) handleMetadataService(ctx context.Context, owner string, vm *domain.VM) error {
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

	vm.Spec.NetworkConfiguration.Interfaces["eth1"] = *metadataInt

	return nil
}

func (a *app) handleNetwork(ctx context.Context, owner string, vm *domain.VM) error {
	pterm.DefaultSpinner.Info("ℹ️  Setting up network")

	bridgeExists, err := a.networkService.BridgeExists(vm.Spec.NetworkConfiguration.BridgeName)
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

	vm.Status.NetworkStatus = map[string]domain.NetworkStatus{}
	for name, intCfg := range vm.Spec.NetworkConfiguration.Interfaces {
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
			if attachErr := a.networkService.AttachToBridge(ifaceName, vm.Spec.NetworkConfiguration.BridgeName); attachErr != nil {
				return fmt.Errorf("attching vm interface to bridge: %w", attachErr)
			}
		}

		vm.Status.NetworkStatus[name] = domain.NetworkStatus{
			HostDeviveName: ifaceName,
			GuestMAC:       mac.ToString(),
		}
	}

	return nil
}

func (a *app) handleMetadata(ctx context.Context, owner string, vm *domain.VM) error {
	networkConfig, err := generateNetworkConfig(vm)
	if err != nil {
		return fmt.Errorf("generating network config")
	}
	vm.Status.Metadata = map[string]string{
		cloudinit.NetworkConfigDataKey: networkConfig,
	}

	metadata, err := a.createMetadata(vm)
	if err != nil {
		return fmt.Errorf("generating metada data: %w", err)
	}
	vm.Status.Metadata[cloudinit.InstanceDataKey] = metadata

	if vm.Spec.Bootstrap != nil {

		userdata, err := a.createUserData(vm)
		if err != nil {
			return fmt.Errorf("generating user data: %w", err)
		}

		vm.Status.Metadata[cloudinit.UserdataKey] = userdata
	}

	return nil

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
