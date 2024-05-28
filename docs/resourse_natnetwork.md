# NAT Networks

## Description
This part enables managing NAT networks in VirtualBox. It allows users to create, read, update, and delete NAT networks along with their associated port forwarding rules. NAT (Network Address Translation) networks in VirtualBox provide a way for virtual machines to communicate with each other and external network using host system's network connection.

## Usage

### Prerequisites

Before using, ensure following prerequisites are met:
- VirtualBox is installed on host system.
- Terraform is installed on host system.

### Configuration

```hcl
resource "virtualbox_natnetwork" "example_nat" {
  name     = "example_nat_network"
  network  = "192.168.10.0/24"
  enabled  = true
  dhcp     = true
  ipv6     = false

  port_forwarding_4 {
    name      = "ssh"
    protocol  = "tcp"
    hostport  = 2222
    guestport = 22
  }
}
```

## Resource 
The resource schema defines configuration options for managing NAT networks. It includes parameters such as `name`, `network`, `enabled`, `dhcp`, `ipv6`, and `port_forwarding`.

## Parameters
- `name` name of NAT network.
- `network` static or DHCP network address and mask of NAT service interface.
- `enabled` enables or disables NAT network service. (Default: true)
- `dhcp` enables or disables DHCP server. (Default: true)
- `ipv6` enables or disables IPv6. (Default: false)
- `port_forwarding_4` list of IPv4 port forwarding rules.
- `port_forwarding_6` list of IPv6 port forwarding rules.
