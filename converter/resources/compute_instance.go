// Package resources implements functions for creating resources in the configuration file for Yandex Cloud.
package resources

import (
	"converter/tools"
	"fmt"
	"os"
	"strconv"
)

// VM_creation creates "yandex_compute_instance".
// The function accepts name of the VM, resources of the VM, network adapters information for each VM,
// the configuration file to write to it, map with OS images for Yandex Cloud, the name of the subnet used and returns any error encountered.
func VM_creation(reference_name string, resources map[string]string, network_adapters map[string](map[int](map[string]string)),
	new_config *os.File, yc_images map[string]string, subnet_name string) error {
	fmt.Fprintf(new_config, "resource \"yandex_compute_instance\" %s {\n", reference_name)

	if err := vm_optional_fields(resources, new_config); err != nil {
		fmt.Println("Error with optional fields of vm.")
		return err
	}

	if err := vm_resource_boot_disk(resources, new_config, yc_images); err != nil {
		fmt.Println("Error with resource boot disk.")
		return err
	}

	vm_resource_network_interface(network_adapters[reference_name], new_config, subnet_name)

	if err := vm_required_resources(resources, new_config); err != nil {
		fmt.Println("Error with required resources of vm.")
		return err
	}

	fmt.Fprint(new_config, "}\n\n")
	return nil
}

// vm_optional_fields converts optional fields such as the number of such machines, the name of the VM and the folder that the VM belongs to.
// The function accepts resources of the VM and the configuration file to write to it and returns any error encountered.
func vm_optional_fields(resources map[string]string, new_config *os.File) error {
	count, ok := resources["count"]
	if ok {
		fmt.Fprintf(new_config, "\tcount = %s\n", count)
	}

	name, ok := resources["name"]
	if ok {
		ok, err := tools.Check_name(name[1 : len(name)-1])
		if ok {
			fmt.Fprintf(new_config, "\tname = %s\n", name)
		} else {
			return err
		}
	}

	folder_name, ok := resources["basedir"]
	if ok {
		fmt.Fprintf(new_config, "\tfolder_id = \"${yandex_resourcemanager_folder.%s.id}\"\n", folder_name[1:len(folder_name)-1])
	}
	return nil
}

// vm_resource_network_interface creates a network interface for VM.
// The function accepts network adapters information for the VM, the configuration file to write to it, the name of the subnet used.
func vm_resource_network_interface(network_adapters map[int]map[string]string, new_config *os.File, subnet_name string) {
	nat := false
	if len(network_adapters) == 0 {
		nat = true
	} else {
		for _, network_adapter := range network_adapters {
			if network_adapter["network_mode"] == "\"nat\"" {
				nat = true
				break
			}
		}
	}

	fmt.Fprint(new_config, "\tnetwork_interface {\n")
	fmt.Fprintf(new_config, "\t\tsubnet_id = \"${yandex_vpc_subnet.%s.id}\"\n", subnet_name)
	if nat {
		fmt.Fprint(new_config, "\t\tnat = true\n")
	}
	fmt.Fprint(new_config, "\t}\n")
}

// vm_required_resources converts the number of processors and memory for the VM.
// The function accepts resources of the VM and the configuration file to write to it and returns any error encountered.
func vm_required_resources(resources map[string]string, new_config *os.File) error {
	fmt.Fprint(new_config, "\tresources {\n")

	str_cpus, ok := resources["cpus"]
	if !ok {
		return tools.ErrCpusNotDefined
	}

	cpus, err := strconv.Atoi(str_cpus)
	if err != nil {
		return err
	}

	//By default, the guaranteed vCPU share is 100%. In this mode, the number of cores and memory must be a multiple of 2.
	if cpus%2 != 0 {
		cpus++
	}

	fmt.Fprintf(new_config, "\t\tcores = %d\n", cpus)

	memory, ok := resources["memory"]
	if !ok {
		return tools.ErrMemoryNotDefined
	}

	mem, err := strconv.Atoi(memory)
	if err != nil {
		return err
	}

	mem /= 1000
	if mem == 0 {
		mem = 2
	} else if mem%2 != 0 {
		mem++
	}

	fmt.Fprintf(new_config, "\t\tmemory = %d\n", mem)
	fmt.Fprint(new_config, "\t}\n")

	return nil
}

// vm_resource_boot_disk converts the system name from the format for VirtualBox to the format for Yandex Cloud.
// The function accepts resources of the VM and the configuration file to write to it,
// map with OS images for Yandex Cloud and returns any error encountered.
func vm_resource_boot_disk(resources map[string]string, new_config *os.File, yc_images map[string]string) error {
	fmt.Fprint(new_config, "\tboot_disk {\n")
	fmt.Fprint(new_config, "\t\tinitialize_params {\n")

	vb_os_id, ok := resources["os_id"]
	if !ok {
		return tools.ErrOsIdNotDefined
	}

	os_name := tools.Get_os_name(vb_os_id)

	yc_os_id, err := tools.Get_yc_os_id(yc_images, os_name)
	if err == nil {
		fmt.Fprintf(new_config, "\t\t\timage_id = \"%s\"\n", yc_os_id)
	} else {
		return err
	}

	fmt.Fprint(new_config, "\t\t}\n")
	fmt.Fprint(new_config, "\t}\n")

	return nil
}
