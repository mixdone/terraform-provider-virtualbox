# Resource_server Documentation

## Description
The resourceVM resource represents a virtual machine that can be managed within your infrastructure. It allows you to define various properties of the virtual machine such as its name, memory allocation, CPU count, image path, network configuration, and more.

## Schema
- `name` (Required): The name of the virtual machine.
- `basedir` (Optional): The folder in which the virtual machine data will be located. Default value is "VMs".
- `memory` (Optional): The amount of RAM allocated for the virtual machine. Default value is 128 MB.
- `vdi_size` (Optional): The size of the VDI (Virtual Disk Image) in MB. Default value is 15000 MB.
- `group` (Optional): The group to which the virtual machine belongs. Default value is an empty string.
- `cpus` (Optional): The number of CPUs allocated to the virtual machine. Default value is 2.
- `status` (Optional): The status of the virtual machine. Default value is "poweroff".
- `image` (Optional): The path to the image located on the host. This property is required when creating a new virtual machine.
- `url` (Optional): The link from which the image or disk will be downloaded. This property is required when creating a new virtual machine.
- `network_adapter` (Optional): Configuration for the network adapter of the virtual machine, including network mode, NIC type, cable connection status, and port forwarding settings.
- `user_data` (Optional): Custom data to be passed to the virtual machine.
- `os_id` (Optional): Specifies the guest OS to run in the VM. It is of type string, and has a default value of "Linux_64".
- `snapshot`: Allows adding a list of snapshots with attributes name (required) and description (optional with a default value of ""). This attribute enables adding, editing, or deleting snapshots for the VM.

## Network Adapter Configuration
The network_adapter property allows you to define the network configuration for the virtual machine. It includes the following sub-properties:
- `index`: The index of the network adapter (computed automatically).
- `network_mode`: The network mode for the adapter (e.g., nat, hostonly). Default value is "none".
- `nic_type`: The type of NIC (Network Interface Controller). Default value is "Am79C970A".
- `cable_connected`: Specifies whether the network cable is connected. Default value is false.
- `port_forwarding`: Configuration for port forwarding, including name, protocol, host IP, host port, guest IP, and guest port.
  


## Example Usage
```hcl
resource "virtualbox_server" "example_vm" {
  name = "my-vm"
  basedir = "VMs"
  memory = 256
  vdi_size = 20000
  group = "my-group"
  cpus = 4
  status = "running"
  image = "/path/to/image.vdi"
  url = "http://example.com/image.iso"

  network_adapter {
    network_mode = "nat"
    nic_type = "Am79C970A"
    cable_connected = true

    port_forwarding {
      name = "ssh"
      protocol = "tcp"
      hostport = 2222
      guestport = 22
    }
  }

  user_data = "#cloud-config\\nhostname: my-vm\\n"
}
```