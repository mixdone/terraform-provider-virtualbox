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
