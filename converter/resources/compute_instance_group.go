// Package resources implements functions for creating resources in the configuration file for Yandex Cloud.
package resources

import (
	"converter/tools"
	"fmt"
	"os"
	"slices"
	"strconv"
)

// Group_creation creates "yandex_compute_instance_group".
// The function accepts general info about VMs, network adapters information for each VM, information about VM groups, the configuration file to write to it,
// map with OS images for Yandex Cloud, the name of the network used, the name of the subnet used, the structure with credentials.
// The function returns any error encountered.
func Group_creation(general_info map[string]map[string]string, network_adapters map[string](map[int](map[string]string)),
	vm_groups map[string][]string, new_config *os.File, yc_images map[string]string, network_name string, subnet_name string, personal_info tools.Info) error {
	for group_name, machines := range vm_groups {
		ok, err := tools.Check_name(group_name[1 : len(group_name)-1])
		if !ok {
			return err
		}

		fmt.Fprintf(new_config, "resource \"yandex_compute_instance_group\" %s {\n", group_name)
		fmt.Fprintf(new_config, "\tname = %s\n", group_name)
		fmt.Fprintf(new_config, "\tfolder_id = \"%s\"\n", personal_info.Folder_ID)
		fmt.Printf("Please write the ID of the service account authorized for %s instance group.\n", group_name)

		var service_account_id string
		if _, err := fmt.Scan(&service_account_id); err != nil {
			return err
		}
		fmt.Fprintf(new_config, "\tservice_account_id = \"%s\"\n", service_account_id)

		fmt.Fprintln(new_config, "\tinstance_template {")
		if err := group_required_resources(machines, general_info, new_config); err != nil {
			return err
		}
		if err := group_resource_boot_disk(machines, general_info, new_config, yc_images); err != nil {
			return err
		}
		group_resource_network_interface(machines, network_adapters, new_config, network_name, subnet_name)
		fmt.Fprintln(new_config, "\t}")

		fmt.Fprintln(new_config, "\tscale_policy {")
		fmt.Fprintln(new_config, "\t\tfixed_scale {")
		size, err := group_size(general_info, machines)
		if err != nil {
			return err
		}
		fmt.Fprintf(new_config, "\t\t\tsize = %d\n", size)
		fmt.Fprintln(new_config, "\t\t}")
		fmt.Fprintln(new_config, "\t}")

		fmt.Fprintln(new_config, "\tallocation_policy {")
		fmt.Fprintf(new_config, "\t\tzones = [\"%s\"]\n", personal_info.Zone)
		fmt.Fprintln(new_config, "\t}")

		fmt.Fprintln(new_config, "\tdeploy_policy {\n\t\tmax_unavailable = 1\n\t\tmax_expansion = 1\n\t}")
		fmt.Fprint(new_config, "}\n\n")
	}
	return nil
}

// group_required_resources finds the number of processors and memory for the VM template for the group and converts them.
// The function accepts array of names of VM belonging to this group, general info about VMs, the configuration file to write to it
// and returns any error encountered.
func group_required_resources(machines []string, general_info map[string]map[string]string, new_config *os.File) error {
	max_memory, max_cpus := 0, 0
	fmt.Fprintln(new_config, "\t\tresources {")
	for _, vm := range machines {
		str_cpus, ok := general_info[vm]["cpus"]
		if !ok {
			return tools.ErrCpusNotDefined
		}
		cpus, err := strconv.Atoi(str_cpus)
		if err != nil {
			return err
		}
		if cpus > max_cpus {
			max_cpus = cpus
		}
		if max_cpus%2 != 0 {
			max_cpus++
		}

		str_memory, ok := general_info[vm]["memory"]
		if !ok {
			return tools.ErrMemoryNotDefined
		}
		memory, err := strconv.Atoi(str_memory)
		if err != nil {
			return err
		}
		memory /= 1000
		if memory == 0 {
			memory++
		}
		if memory > max_memory {
			max_memory = memory
		}
		if max_memory%2 != 0 {
			max_memory++
		}
	}

	fmt.Fprintf(new_config, "\t\t\tmemory = %d\n", max_memory)
	fmt.Fprintf(new_config, "\t\t\tcores = %d\n", max_cpus)
	fmt.Fprintln(new_config, "\t\t}")
	return nil
}

// group_resource_boot_disk selects the operating system for the VM template for the group based on the user's choice
// and converts the system name from the format for VirtualBox to the format for Yandex Cloud.
// The function accepts array of names of VM belonging to this group, general info about VMs, the configuration file to write to it,
// map with OS images for Yandex Cloud and returns any error encountered.
func group_resource_boot_disk(machines []string, general_info map[string]map[string]string, new_config *os.File, yc_images map[string]string) error {
	fmt.Fprint(new_config, "\t\tboot_disk {\n")
	fmt.Fprint(new_config, "\t\t\tinitialize_params {\n")
	systems := make([]string, 0)
	for _, vm := range machines {
		vb_os_id, ok := general_info[vm]["os_id"]
		if !ok {
			return tools.ErrOsIdNotDefined
		}
		if !slices.Contains(systems, vb_os_id) {
			systems = append(systems, vb_os_id)
		}
	}

	var system string
	var err error
	if len(systems) == 1 {
		system = systems[0]
	} else {
		for {
			fmt.Println("Please select the os for group of vm from this list. Write only the number")
			system, err = tools.Choose_option(systems)
			if err != tools.ErrInvalidInput {
				break
			}
		}
	}

	os_name := tools.Get_os_name(system)

	yc_os_id, err := tools.Get_yc_os_id(yc_images, os_name)
	if err == nil {
		fmt.Fprintf(new_config, "\t\t\t\timage_id = \"%s\"\n", yc_os_id)
	} else {
		return err
	}

	fmt.Fprint(new_config, "\t\t\t}\n")
	fmt.Fprint(new_config, "\t\t}\n")

	return nil
}

// group_resource_network_interface creates a network interface for VMs group.
// The function accepts array of names of VM belonging to this group, network adapters information for each VM,
// the configuration file to write to it, the name of the network used, the name of the subnet used.
func group_resource_network_interface(machines []string, network_adapters map[string](map[int](map[string]string)), new_config *os.File,
	network_name string, subnet_name string) {
	nat := false
	for _, vm := range machines {
		if len(network_adapters[vm]) == 0 {
			nat = true
			break
		} else {
			for _, network_adapter := range network_adapters[vm] {
				if network_adapter["network_mode"] == "\"nat\"" {
					nat = true
					break
				}
			}
			if nat {
				break
			}
		}
	}

	fmt.Fprint(new_config, "\t\tnetwork_interface {\n")
	fmt.Fprintf(new_config, "\t\t\tnetwork_id = \"${yandex_vpc_network.%s.id}\"\n", network_name)
	fmt.Fprintf(new_config, "\t\t\tsubnet_ids = [\"${yandex_vpc_subnet.%s.id}\"]\n", subnet_name)
	if nat {
		fmt.Fprint(new_config, "\t\t\tnat = true\n")
	}
	fmt.Fprint(new_config, "\t\t}\n")
}

// group_size calculates the number of VMs in a group.
// The function accepts general info about VMs, array of names of VM belonging to this group
// and returns the number of VMs in a group and any error encountered.
func group_size(general_info map[string]map[string]string, machines []string) (int, error) {
	size := 0
	for _, vm := range machines {
		str_count, ok := general_info[vm]["count"]
		if ok {
			count, err := strconv.Atoi(str_count)
			if err != nil {
				return size, err
			}
			size += count
		} else {
			size++
		}
	}
	return size, nil
}
