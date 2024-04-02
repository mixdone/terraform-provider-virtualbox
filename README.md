<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
    <img src="https://raw.githubusercontent.com/mixdone/terraform-provider-virtualbox/main/assets/terraform-logo.png" alt="Terraform logo" title="Terraform" align="right" height="75" />
</a>
<a href="https://www.virtualbox.org/">
    <img src="https://raw.githubusercontent.com/mixdone/terraform-provider-virtualbox/main/assets/vb-logo.png" alt="VirtualBox logo" title="VirtualBox" align="right" height="75" />
</a>

# terraform-provider-virtualbox

The [Terraform Provider](https://registry.terraform.io/providers/daria-barsukova/virtualbox/latest) allows [Terraform](https://terraform.io) to manage [VirtualBox](https://www.virtualbox.org/) resources.

[![Release](https://img.shields.io/github/v/release/daria-barsukova/terraform-provider-virtualbox)](https://github.com/daria-barsukova/terraform-provider-virtualbox/releases)
[![Installs](https://img.shields.io/badge/dynamic/json?logo=terraform&label=installs&query=$.data.attributes.downloads&url=https%3A%2F%2Fregistry.terraform.io%2Fv2%2Fproviders%2F712)](https://registry.terraform.io/providers/daria-barsukova/virtualbox)
[![Registry](https://img.shields.io/badge/registry-doc%40latest-lightgrey?logo=terraform)](https://registry.terraform.io/providers/daria-barsukova/virtualbox/latest/docs)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](https://github.com/mixdone/terraform-provider-virtualbox/blob/main/LICENSE)  
[![Go Status](https://github.com/mixdone/terraform-provider-virtualbox/workflows/CI/badge.svg)](https://github.com/mixdone/terraform-provider-virtualbox/actions)
[![Lint Status](https://github.com/mixdone/terraform-provider-virtualbox/workflows/CodeQL/badge.svg)](https://github.com/mixdone/terraform-provider-virtualbox/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/mixdone/terraform-provider-virtualbox)](https://goreportcard.com/report/github.com/mixdone/terraform-provider-virtualbox)  


## Example usage

Take a look at the examples in the [documentation](https://registry.terraform.io/providers/daria-barsukova/virtualbox/latest/docs) of the registry or use the following example:

```hcl
# Creating a resource "virtualbox_server" with the name "VM_without_image"
resource "virtualbox_server" "VM_without_image" {
    count     = 0
    name      = format("VM_without_image-%02d", count.index + 1)   # Formatting the name of the virtual machine
    basedir = format("VM_without_image-%02d", count.index + 1)     # Formatting the base directory for the virtual machine
    cpus      = 3                                                  # Setting the number of virtual CPUs
    memory    = 1000                                               # Setting the memory size for the virtual machine
    status = "running"                                             # Setting the status of the virtual machine to "running"
    os_id = "Windows7_64"                                          # Setting the operating system identifier
}

# Creating a resource "virtualbox_server" with the name "bad_VM_example"
resource "virtualbox_server" "bad_VM_example" {
    count     = 0
    name      = format("VM_without_image-%02d", count.index + 1)   # Formatting the name of the virtual machine
    basedir = format("VM_without_image-%02d", count.index + 1)     # Formatting the base directory for the virtual machine
    cpus      = 3                                                  # Setting the number of virtual CPUs
    memory    = 2500                                               # Setting a higher memory size for the virtual machine
    status = "poweroff"                                            # Setting the status of the virtual machine to "poweroff"
    os_id = "Windows7_64"                                          # Setting the operating system identifier
    group = "/man"                                                 # Assigning the virtual machine to a specific group

    snapshot {                                                     # Creating a snapshot for the virtual machine
      name = "hello"                                               # Setting the name of the snapshot
      description = "hohohhoho"                                    # Providing a description for the snapshot
    }
}
```

## Guide for launching the provider on your device 

1. Download the contents of the main branch to your device

2. Build the provider code using the command
  * Linux
    ```
    go build -o terraform-provider-virtualbox
    ```
    
  * Windows
    ```
    go build -o terraform-provider-virtualbox.exe
    ```

3. In order to use the provider, we need to create the below directory structure inside the plugins directory 
  * Linux
    ```
    ~/.terraform.d/plugins/${host_name}/${namespace}/${type}/${version}/${target}
    ```

  * Windows
    ```
    %APPDATA%\terraform.d\plugins\${host_name}\${namespace}\${type}\${version}\${target}
    ```
    Where:
      * host_name -> somehostname.com
      * namespace -> provider name space
      * type -> provider type
      * version -> semantic versioning of the provider
      * target -> target operating system

  * As a first step, we need to the create the directory
    * Linux
    ```
    mkdir -p ~/.terraform.d/plugins/terraform-virtualbox.local/virtualboxprovider/virtualbox/1.0.0/linux_amd64
    ```
    
    * Windows
    ```
    mkdir Path_to_the_AppData_folder\AppData\Roaming\terraform.d\plugins\terraform-virtualbox.local\virtualboxprovider\virtualbox\1.0.0\windows_amd64
    ```
    
    * MacOS
    ```
    mkdir Path_to_the_AppData_folder\AppData\Roaming\terraform.d\plugins\terraform-virtualbox.local\virtualboxprovider\virtualbox\1.0.0\darwin_x86_64
    ```
    
  * Then, copy the terraform-provider-virtualbox to the created directory
    * Linux
    ```
    cp terraform-provider-virtualbox ~/.terraform.d/plugins/terraform-virtualbox.local/virtualboxprovider/virtualbox/1.0.0/linux_amd64
    ```
    
    * Windows
    ```
    cp terraform-provider-virtualbox.exe Path_to_the_AppData_folder\AppData\Roaming\terraform.d\plugins\terraform-virtualbox.local\virtualboxprovider\virtualbox\1.0.0\windows_amd64
    ```
    
    * MacOS
    ```
    cp terraform-provider-virtualbox.exe Path_to_the_AppData_folder\AppData\Roaming\terraform.d\plugins\terraform-virtualbox.local\virtualboxprovider\virtualbox\1.0.0\darwin_x86_64
    ```
    
4. Go to the config folder
```
cd examples/resources
```

5. Use the commands 
  * Terraform init
  * Terraform plan
  * Terraform apply
  * Terraform destroy

## Support
For any issues or questions related to this provider, please open an issue on the [GitHub repository](https://github.com/mixdone/terraform-provider-virtualbox)

## License

The Terraform Provider VirtualBox is available to everyone under the terms of the Apache Public License Version 2.0. [Take a look the LICENSE file](LICENSE).
