# mikrolite - the microvm cli

**mikrolite** is a tool that can be used to create and manage microvms from the command line. Its aim is to make working with microvms easy from the command line.

> NOTE: this is very much a prototype.

## What are microvms?

In this instance, the term **microvm** refers to lightweight virtualisation. And more specifically 2 implementions:

- [Firecracker](https://firecracker-microvm.github.io/)
- [Cloud Hypervisor](https://www.cloudhypervisor.org/)

Both of these are supported by mikrolite.

## Requirements

- Linux host machine with kvm
- containerd with the devmapper snapshotter plugin enabled & configured
- Either or both of these:
  - Firecracker (tested with v1.5.0)
  - Cloud Hypervisor (tested with v36)

This is an example of the configuration section that will need to be added to your container config:

```toml
[plugins]
  [plugins."io.containerd.snapshotter.v1.devmapper"]
    pool_name = "development-thinpool"
    root_path = "/var/lib/containerd-dev/snapshotter/devmapper"
    base_image_size = "10GB"
    discard_blocks = true
```

## Very basic usage

To create a VM using Firecracker

```shell
sudo ./mikrolite vm create --name node1 --root-image ghcr.io/mikrolite/node-rke2-airgapped:dev --kernel-image ghcr.io/mikrolite/firecracker-kernel:5.10 --kernel-filename boot/vmlinux --provider firecracker --firecracker-bin /path/to/firecracker-v1.5.0-x86_64 --network-bridge virbr0 --ssh-key /home/user/.ssh/id_ed25519.pub
```

> Currently the bridge needs to exist before running the **create** command. The easiest way to do this is using **virt-manager** or **virsh** as this will setup iptables etc.

After the VM boots you should be able to connect to the vm via SSH:

```shell
ssh ml@<IP_OF_VM>
```

To get a list of vms:

```shell
sudo ./mikrolite vm list
```

## Contributing

We'd love your help on this via issues, PRs etc.
