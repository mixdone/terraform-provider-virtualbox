<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
    <img src="https://raw.githubusercontent.com/mixdone/terraform-provider-virtualbox/main/assets/terraform-logo.png" alt="Terraform logo" title="Terraform" align="right" height="75" />
</a>
<a href="https://www.virtualbox.org/">
    <img src="https://raw.githubusercontent.com/mixdone/terraform-provider-virtualbox/main/assets/vb-logo.png" alt="VirtualBox logo" title="VirtualBox" align="right" height="75" />
</a>

# terraform-provider-virtualbox

[![Release](https://img.shields.io/github/v/release/daria-barsukova/terraform-provider-virtualbox)](https://github.com/daria-barsukova/terraform-provider-virtualbox/releases)
[![Installs](https://img.shields.io/badge/dynamic/json?logo=terraform&label=installs&query=$.data.attributes.downloads&url=https%3A%2F%2Fregistry.terraform.io%2Fv2%2Fproviders%2F712)](https://registry.terraform.io/providers/daria-barsukova/virtualbox)
[![Registry](https://img.shields.io/badge/registry-doc%40latest-lightgrey?logo=terraform)](https://registry.terraform.io/providers/daria-barsukova/virtualbox/latest/docs)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](https://github.com/mixdone/terraform-provider-virtualbox/blob/main/LICENSE)  
[![Go Status](https://github.com/mixdone/terraform-provider-virtualbox/workflows/CI/badge.svg)](https://github.com/mixdone/terraform-provider-virtualbox/actions)
[![Lint Status](https://github.com/mixdone/terraform-provider-virtualbox/workflows/CodeQL/badge.svg)](https://github.com/mixdone/terraform-provider-virtualbox/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/mixdone/terraform-provider-virtualbox)](https://goreportcard.com/report/github.com/mixdone/terraform-provider-virtualbox)  

The [Terraform Provider](https://registry.terraform.io/providers/daria-barsukova/virtualbox/latest) allows [Terraform](https://terraform.io) to manage [VirtualBox](https://www.virtualbox.org/) resources.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html)  v1.5.7+
-	[Go](https://golang.org/doc/install) 1.21.1 (to build the provider plugin)
-  [VirtualBox](https://www.virtualbox.org/manual/ch02.html)

## Provider Capabilities

The provider provides the following features:

1. Creating a virtual machine:
   - Accept input parameters such as the name of the virtual machine, the path to the image, the size of memory, the number of processors and other configuration parameters.
   - Create a folder for the virtual machine in the specified folder.
   - Assign a status to the VM (for example, "running" or "poweroff").
   - Configure the network adapter with a specific operating mode, type of NIC, whether the cable is connected and port forwarding settings.
   - Transfer user data (for example, configuration scripts) inside the virtual machine.

2. Updating the virtual machine:
   - Update the parameters of the virtual machine.
   - Change the configuration of the network adapter, including the operating mode and port settings.
   - Update user data.

3. Deleting a virtual machine:
   - Remove a VM from the infrastructure.
   - Clear the folder with the virtual machine data.

4. Network Settings Management:
   - Add/remove port forwarding settings for the virtual machine.

5. Getting information about a virtual machine:
   - Return information about the current state of the virtual machine, such as status, memory usage, number of processors, and other parameters.

6. Working with images:
   - Upload images from the specified URL.
   - Manage paths to images and their sizes.
  
7. Managing virtual machine snapshots, including creating new snapshots, editing snapshot descriptions, and deleting existing snapshots.

## Example usage

* [Guide for launching the provider on your device ](https://github.com/mixdone/terraform-provider-virtualbox/blob/main/GUIDE.md)

* Take a look at the examples in the [documentation](https://registry.terraform.io/providers/daria-barsukova/virtualbox/latest/docs) of the registry or use the following example:

```hcl
# Define a VirtualBox server resource for creating VMs with network configurations
resource "virtualbox_server" "VM_network" {
  count   = 0
  name    = format("VM_network-%02d", count.index + 1)  # Name of the VM
  basedir = format("VM_network-%02d", count.index + 1)  # Base directory for VM files
  cpus    = 3                                           # Number of CPUs for the VM
  memory  = 500                                         # Amount of memory in MB for the VM

  # Network adapter configurations
  network_adapter {
    network_mode = "nat"                                # NAT mode for network adapter
    port_forwarding {
      name      = "rule1"
      hostip    = ""                                    # Host IP address for port forwarding
      hostport  = "80"                                  # Host port for port forwarding
      guestip   = ""                                    # Guest IP address for port forwarding
      guestport = "63222"                               # Guest port for port forwarding
    }
  }
  network_adapter {
    network_mode    = "nat"                             # NAT mode for network adapter
    nic_type        = "82540EM"                         # Type of network interface controller
    cable_connected = true                              # Whether the cable is connected
  }
  network_adapter {
    network_mode = "hostonly"                          # Host-only mode for network adapter
  }
  network_adapter {
    network_mode = "bridged"                            # Bridged mode for network adapter
    nic_type     = "virtio"                             # Type of network interface controller
  }

  status = "poweroff"                                   # Initial status of the VM
}

# Define a VirtualBox server resource for creating VMs with snapshots
resource "virtualbox_server" "VM_Shapshots" {
  count   = 0
  name    = format("VM_Snapshots-%02d", count.index + 1)  # Name of the VM
  basedir = format("VM_Snapshots-%02d", count.index + 1)  # Base directory for VM files
  cpus    = 4                                              # Number of CPUs for the VM
  memory  = 2000                                           # Amount of memory in MB for the VM

  # Define snapshots for the VM
  snapshot {
    name        = "first"                                  # Name of the snapshot
    description = "example"                                # Description of the snapshot
  }

  snapshot {
    name     = "second"                                    # Name of the snapshot
    description = "example"                                # Description of the snapshot
    current  = true                                        # Set this snapshot as current
  }
}
```

## Support
For any issues or questions related to this provider, please open an issue on the [GitHub repository](https://github.com/mixdone/terraform-provider-virtualbox)

## License

The Terraform Provider VirtualBox is available to everyone under the terms of the Apache Public License Version 2.0. [Take a look the LICENSE file](LICENSE).
