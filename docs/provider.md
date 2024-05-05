# Provider Documentation

## Description
This Terraform provider allows you to manage VirtualBox virtual machines. It provides resources to create and configure virtual machines within VirtualBox.

## Usage
To use this provider, you need to have VirtualBox installed on your machine. You can specify the configuration for virtual machines using the resources provided by this provider.

## Resources
### virtualbox_server
The `virtualbox_server` resource allows you to create and manage VirtualBox virtual machines. It supports the following arguments:
- `name` (string): The name of the virtual machine.
- `basedir` (string): The base directory for the virtual machine.
- `cpus` (int): The number of virtual CPUs for the virtual machine.
- `memory` (int): The memory size for the virtual machine.
- `status` (string): The status of the virtual machine (e.g., "running", "poweroff").
- `os_id` (string): The operating system identifier for the virtual machine.

#### Example
```hcl
resource "virtualbox_server" "example_vm" {
  name      = "my_vm"
  basedir   = "/path/to/vm/directory"
  cpus      = 2
  memory    = 2048
  status    = "running"
  os_id     = "Ubuntu_64"
}
```

## Provider Configuration
In Terraform configuration file you need to specify the provider block to use this provider:
```hcl
terraform {
  required_providers {
    virtualbox = {
      version = "~> 1.0.0"
      source  = "terraform-virtualbox.local/virtualboxprovider/virtualbox"
    }
  }
}
```

## Version Compatibility
This provider is compatible with VirtualBox version 6.0 and above.

