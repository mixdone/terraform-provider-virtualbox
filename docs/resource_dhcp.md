# DHCP

## Description
It provides a Terraform provider for managing DHCP servers using VirtualBox. It allows users to create, read, update, and delete DHCP server configurations.

## Usage

To use it, include in your Terraform configuration. For example:

```hcl
resource "dhcp_server" "example" {
  server_ip     = "192.168.1.1"
  lower_ip      = "192.168.1.100"
  upper_ip      = "192.168.1.200"
  network_name  = "example_network"
  network_mask  = "255.255.255.0"
  enabled       = true
}
```
## Resources
The DHCP server resource supports the following attributes:

- `server_ip` IP address of DHCP server.
- `lower_ip` lower bound for IP addresses to assign.
- `upper_ip` upper bound for IP addresses to assign.
- `network_name` name of the network where DHCP server will be running.
- `network_mask` network mask.
 - `enabled` boolean indicating whether DHCP server is enabled or disabled.
  
## Resource Operations
- Create
This operation retrieves DHCP configuration parameters from Terraform resource data, creates DHCP server using VirtualBox API, and stores DHCP server's ID.

- Read
The Read operation retrieves DHCP server configuration parameters from VirtualBox API and sets them in the Terraform resource data.

- Update: compares old and new DHCP configurations, modifies DHCP server accordingly, and updates configuration using VirtualBox API.

- Delete: retrieves DHCP server configuration and removes it using VirtualBox API.

- Exists: checks if DHCP server configuration exists by verifying its existence.

- Error handling: handled gracefully throughout provider. Diagnostics are used to provide meaningful error messages to users in case of failures during resource operations.

- Dependencies: provider depends on virtualbox-go package for interacting with VirtualBox via its API.
