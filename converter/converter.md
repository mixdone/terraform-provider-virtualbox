# Converter
This converter is designed to convert a configuration file for terraform-provider-virtualbox into a configuration file for Yandex Cloud provider.
## Guide to using the converter
1. You need to put a configuration file of the format *.tf* that you want to convert to the same directory where the converter.go file is located and run the converter on the command line using the command.
    ```
    go run converter.go
    ```
2. Next, you will need to select the appropriate file for conversion from the list of suggested ones.
3. The provider needs to be configured with the proper credentials before it can be used. To do this, you need to answer a few questions about the [IAM token](https://yandex.cloud/en/docs/iam/operations/iam-token/create "About IAM token"), the [cloud ID](https://yandex.cloud/ru/docs/resource-manager/operations/cloud/get-id#console_1 "About cloud ID"), the default [folder ID](https://yandex.cloud/ru/docs/resource-manager/operations/folder/get-id#api_1 "About folder ID"), and the [default availability zone](https://yandex.cloud/en/docs/overview/concepts/geo-scope "About availability zones").
4. Then you need to answer the questions to fill in those fields that are not used in the original configuration file, but are necessary for the configuration file for the Yandex cloud.
### Validity of names
The names of resources such as networks, subnets, virtual machines, folders, and virtual machine groups must satisfy the **following requirement**:\
The name can contain lowercase Latin letters, numbers, and hyphens. The first character must be a letter. The last character must not be a hyphen. The allowed length is from 2 to 63 characters.
### "yandex_vpc_network" and "yandex_vpc_subnet"
* Networks and subnets are created in the default folder. 
* The default availability zone is used for the subnet.
### "yandex_compute_instance"
* The virtual machine is created in the folder that is specified in the configuration file for the provider. If the folder is not specified, then the VM is created in the default folder.
* The default availability zone is used for the VM.
* By default, the guaranteed vCPU share is 100%. In this mode, the number of cores and memory must be a multiple of 2. Therefore, the number of processors and memory is rounded up to an even number.
### "yandex_compute_instance_group"
* The virtual machine group is created in the default folder.
* The default availability zone is used for the virtual machine group.
* As an answer to one of the questions, you need to enter the ID of the [service account](https://yandex.cloud/ru/docs/iam/concepts/users/service-accounts "About service accounts") authorized for this instance group. To be able to create, update, and delete VMs in a group, assign **the editor role** to the service account.
* In VirtualBox, you can combine machines with different parameters into a group, but in Yandex cloud, all machines in the group must be the same, so a template is created for them. To create a **virtual machine template**, all machines belonging to this group in the configuration file for VirtualBox are analyzed.
    - The number of processors and memory for the VM template is the maximum number of processors and memory for all machines in this group.
    - The user chooses the operating system for the VM template by answering the question.
* The **deploy policy** for the group is set automatically.
    - The maximum number of running instances that can be taken offline (stopped or deleted) at the same time
during the update process (max_unavailable) is 1.
    - The maximum number of instances that can be temporarily allocated above the group's target size
during the update process (max_expansion) is 1.

