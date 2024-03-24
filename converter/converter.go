package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var ErrInvalidOperatingSystemName = errors.New("finding suitable yc_os_id failed")
var ErrNoConfigurationFile = errors.New("there is no configuration file for conversion in the directory")
var ErrOsIdNotDefined = errors.New("os_id must be specified")
var ErrCpusNotDefined = errors.New("cpus must be specified")
var ErrMemoryNotDefined = errors.New("memory must be specified")
var ErrInvalidInput = errors.New("invalid input")
var ErrInvalidName = errors.New("invalid resource name")

type info struct {
	OAuth_token string
	cloud_ID    string
	folder_ID   string
	zone        string
}

func scan_info(config *os.File) (map[string](map[string]string), map[string](map[int](map[string]string)), map[string][]string, error) {
	general_info := make(map[string](map[string]string))
	network_adapters := make(map[string](map[int](map[string]string)))
	vm_groups := make(map[string][]string)
	scanner := bufio.NewScanner(config)
	currVM := ""
	network_adapter_count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			splited := strings.Split(line, "=")
			general_info[currVM][strings.TrimSpace(splited[0])] = strings.TrimSpace(splited[1])
			if strings.TrimSpace(splited[0]) == "group" {
				vm_groups[strings.TrimSpace(splited[1])] = append(vm_groups[strings.TrimSpace(splited[1])], currVM)
			}
		} else if strings.Contains(line, "resource \"virtualbox_server\"") {
			splited := strings.Split(line, " ")
			currVM = splited[2]
			general_info[currVM] = make(map[string]string)
			network_adapters[currVM] = make(map[int](map[string]string))
		} else if strings.Contains(line, "network_adapter") {
			network_adapters[currVM][network_adapter_count] = make(map[string]string)
			if scanner.Scan() {
				line = scanner.Text()
			}
			for {
				if strings.TrimSpace(line) == "}" {
					break
				} else if strings.Contains(line, "=") {
					splited := strings.Split(line, "=")
					network_adapters[currVM][network_adapter_count][strings.TrimSpace(splited[0])] = strings.TrimSpace(splited[1])
				}
				if scanner.Scan() {
					line = scanner.Text()
				}
			}
			network_adapter_count++
		}
	}

	err := scanner.Err()

	return general_info, network_adapters, vm_groups, err
}

func get_yc_images() (map[string]string, error) {
	yc_images := make(map[string]string)

	out, err := exec.Command("powershell", "yc compute image list --folder-id standard-images").Output()
	if err != nil {
		return nil, err
	}

	splited_yc_images := strings.Split(string(out), "|")

	for i := 7; i < len(splited_yc_images)-1; i += 6 {
		yc_images[strings.TrimSpace(splited_yc_images[i+1])] = strings.TrimSpace(splited_yc_images[i])
	}

	return yc_images, nil
}

func get_os_name(vb_os_id string) string {
	os_name := ""

	os_id := strings.Split(vb_os_id, "_")[0]

	for i, ch := range os_id {
		if strings.Contains("0123456789", string(ch)) {
			os_name += "-"
			for j := i; j < len(os_id); j++ {
				if string(os_id[j]) != "\"" {
					os_name += string(os_id[j])
				}
			}
			break
		}
		if string(ch) != "\"" {
			os_name += string(ch)
		}
	}

	os_name = strings.ToLower(os_name)

	return os_name
}

func get_yc_os_id(yc_images map[string]string, os_name string) (string, error) {
	yc_os_id := ""

	for name, id := range yc_images {
		if strings.Contains(name, os_name) {
			yc_os_id = id
			break
		}
	}

	if yc_os_id == "" {
		return yc_os_id, ErrInvalidOperatingSystemName
	} else {
		return yc_os_id, nil
	}
}

func select_config() (string, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}

	configs := make([]string, 0, 1)

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".tf" {
			configs = append(configs, file.Name())
		}
	}

	if len(configs) == 0 {
		return "", ErrNoConfigurationFile
	} else {
		var config string
		for {
			fmt.Println("Please select the file to convert from this list. Write only the number.")
			config, err = choose_option(configs)
			if err != ErrInvalidInput {
				break
			}
		}
		fmt.Println(config)
		return config, nil
	}
}

func choose_option(options []string) (string, error) {
	var num int
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	if _, err := fmt.Scan(&num); err != nil {
		fmt.Println("Invalid input!")
		return "", ErrInvalidInput
	}

	if (1 <= num) && (num <= 4) {
		return options[num-1], nil
	}

	fmt.Println("Invalid number!")
	return "", ErrInvalidInput
}

func check_name(name string) (bool, error) {
	re := regexp.MustCompile(`^[a-z]+[a-z0-9\-]+[a-z0-9]+$`)
	if !re.MatchString(name) {
		fmt.Printf(`Invalid name "%s"! The name can contain lowercase Latin letters, numbers, and hyphens. 
The first character must be a letter. The last character must not be a hyphen. 
The allowed length is from 2 to 63 characters.\n`, name)
		return false, ErrInvalidName
	}
	return true, nil
}

func network_creation(new_config *os.File) (string, error) {
	var network_name string
	for {
		fmt.Println("Please write a name for the network.")
		if _, err := fmt.Scan(&network_name); err != nil {
			return "", err
		}

		if ok, _ := check_name(network_name); ok {
			break
		}
	}

	fmt.Fprintf(new_config, "resource \"yandex_vpc_network\" \"%s\" {\n", network_name)
	fmt.Fprintf(new_config, "\tname = \"%s\"\n", network_name)
	fmt.Fprint(new_config, "}\n\n")
	return network_name, nil
}

func subnet_creation(new_config *os.File, network_name string, zone string) (string, error) {
	var subnet_name string
	for {
		fmt.Println("Please write a name for the subnet.")
		_, err := fmt.Scan(&subnet_name)
		if err != nil {
			return "", err
		}

		if ok, _ := check_name(subnet_name); ok {
			break
		}
	}

	fmt.Fprintf(new_config, "resource \"yandex_vpc_subnet\" \"%s\" {\n", subnet_name)
	fmt.Fprintf(new_config, "\tname = \"%s\"\n", subnet_name)

	var v4_cidr_blocks string
	for {
		fmt.Println(`Please write a list of blocks of internal IPv4 addresses that are owned by this subnet.
For example, 10.0.0.0/22 or 192.168.0.0/16.
Blocks of addresses must be unique and non-overlapping within a network.
Minimum subnet size is /28, and maximum subnet size is /16. Only IPv4 is supported.`)
		_, err := fmt.Scan(&v4_cidr_blocks)
		if err != nil {
			return "", err
		}
		splited := strings.Split(v4_cidr_blocks, "/")
		if len(splited) < 2 {
			fmt.Println("Invalid input!")
			continue
		}
		mask, err := strconv.Atoi(splited[1])
		if (err != nil) || (16 > mask) || (mask > 28) {
			fmt.Println("Invalid subnet size!")
			continue
		}
		_, err = netip.ParseAddr(splited[0])
		if err != nil {
			fmt.Println("Invalid IPv4 addresses!")
			continue
		}
		break
	}
	fmt.Fprintf(new_config, "\tv4_cidr_blocks = [\"%s\"]\n", v4_cidr_blocks)

	fmt.Fprintf(new_config, "\tzone = \"%s\"\n", zone)
	fmt.Fprintf(new_config, "\tnetwork_id = \"${yandex_vpc_network.%s.id}\"\n", network_name)
	fmt.Fprint(new_config, "}\n\n")
	return subnet_name, nil
}

func vm_optional_fields(resources map[string]string, new_config *os.File) error {
	count, ok := resources["count"]
	if ok {
		fmt.Fprintf(new_config, "\tcount = %s\n", count)
	}

	name, ok := resources["name"]
	if ok {
		ok, err := check_name(name[1 : len(name)-1])
		if ok {
			fmt.Fprintf(new_config, "\tname = %s\n", name)
		} else {
			return err
		}
	}
	return nil
}

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

func vm_required_resources(resources map[string]string, new_config *os.File) error {
	fmt.Fprint(new_config, "\tresources {\n")

	str_cpus, ok := resources["cpus"]
	if !ok {
		return ErrCpusNotDefined
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
		return ErrMemoryNotDefined
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

func vm_resource_boot_disk(resources map[string]string, new_config *os.File, yc_images map[string]string) error {
	fmt.Fprint(new_config, "\tboot_disk {\n")
	fmt.Fprint(new_config, "\t\tinitialize_params {\n")

	vb_os_id, ok := resources["os_id"]
	if !ok {
		return ErrOsIdNotDefined
	}

	os_name := get_os_name(vb_os_id)

	yc_os_id, err := get_yc_os_id(yc_images, os_name)
	if err == nil {
		fmt.Fprintf(new_config, "\t\t\timage_id = \"%s\"\n", yc_os_id)
	} else {
		return err
	}

	fmt.Fprint(new_config, "\t\t}\n")
	fmt.Fprint(new_config, "\t}\n")

	return nil
}

func group_required_resources(machines []string, general_info map[string]map[string]string, new_config *os.File) error {
	max_memory, max_cpus := 0, 0
	fmt.Fprintln(new_config, "\t\tresources {")
	for _, vm := range machines {
		str_cpus, ok := general_info[vm]["cpus"]
		if !ok {
			return ErrCpusNotDefined
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
			return ErrMemoryNotDefined
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

func group_resource_boot_disk(machines []string, general_info map[string]map[string]string, new_config *os.File, yc_images map[string]string) error {
	fmt.Fprint(new_config, "\t\tboot_disk {\n")
	fmt.Fprint(new_config, "\t\t\tinitialize_params {\n")
	systems := make([]string, 0)
	for _, vm := range machines {
		vb_os_id, ok := general_info[vm]["os_id"]
		if !ok {
			return ErrOsIdNotDefined
		}
		systems = append(systems, vb_os_id)
	}

	var system string
	var err error
	if len(systems) == 1 {
		system = systems[0]
	} else {
		for {
			fmt.Println("Please select the os for group of vm from this list. Write only the number")
			system, err = choose_option(systems)
			if err != ErrInvalidInput {
				break
			}
		}
	}

	os_name := get_os_name(system)

	yc_os_id, err := get_yc_os_id(yc_images, os_name)
	if err == nil {
		fmt.Fprintf(new_config, "\t\t\t\timage_id = \"%s\"\n", yc_os_id)
	} else {
		return err
	}

	fmt.Fprint(new_config, "\t\t\t}\n")
	fmt.Fprint(new_config, "\t\t}\n")

	return nil
}

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

func resource_group(general_info map[string]map[string]string, network_adapters map[string](map[int](map[string]string)),
	vm_groups map[string][]string, new_config *os.File, yc_images map[string]string, network_name string, subnet_name string, personal_info info) error {
	for group_name, machines := range vm_groups {
		ok, err := check_name(group_name[1 : len(group_name)-1])
		if !ok {
			return err
		}
		fmt.Fprintf(new_config, "resource \"yandex_compute_instance_group\" %s {\n", group_name)
		fmt.Fprintf(new_config, "\tname = %s\n", group_name)
		fmt.Fprintf(new_config, "\tfolder_id = \"%s\"\n", personal_info.folder_ID)
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
		fmt.Fprintf(new_config, "\t\tzones = [\"%s\"]\n", personal_info.zone)
		fmt.Fprintln(new_config, "\t}")
		fmt.Fprintln(new_config, "\tdeploy_policy {\n\t\tmax_unavailable = 1\n\t\tmax_expansion = 1\n\t}")
		fmt.Fprintln(new_config, "}\n")
	}
	return nil
}

func resource_vm(reference_name string, resources map[string]string, network_adapters map[string](map[int](map[string]string)),
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

func conversion_of_resources(general_info map[string]map[string]string, network_adapters map[string](map[int](map[string]string)),
	vm_groups map[string][]string, new_config *os.File, yc_images map[string]string, network_name string, subnet_name string, personal_info info) {
	if len(vm_groups) != 0 {
		if err := resource_group(general_info, network_adapters, vm_groups, new_config, yc_images, network_name, subnet_name, personal_info); err != nil {
			new_config.Close()
			os.Remove("yandex_cloud.tf")
			log.Fatalf("Error with resource group: %s", err.Error())
		}
	}
	for reference_name, resources := range general_info {
		if _, ok := resources["group"]; !ok {
			if err := resource_vm(reference_name, resources, network_adapters, new_config, yc_images, subnet_name); err != nil {
				new_config.Close()
				os.Remove("yandex_cloud.tf")
				log.Fatalf("Error with resource virtual machine: %s", err.Error())
			}
		}
	}
}

func get_personal_info(new_config *os.File) (info, error) {
	var personal_info info
	fmt.Fprintln(new_config, "terraform {\n\trequired_providers {\n\t\tyandex = {\n\t\t\tsource = \"yandex-cloud/yandex\"\n\t\t}\n\t}\n}\n")
	fmt.Fprintln(new_config, "provider \"yandex\" {")
	fmt.Println("Please write security token or IAM token used for authentication in Yandex.Cloud.")
	if _, err := fmt.Scan(&personal_info.OAuth_token); err != nil {
		return personal_info, err
	}
	fmt.Fprintf(new_config, "\ttoken = \"%s\"\n", personal_info.OAuth_token)
	fmt.Println("Please write the ID of the cloud to apply any resources to.")
	if _, err := fmt.Scan(&personal_info.cloud_ID); err != nil {
		return personal_info, err
	}
	fmt.Fprintf(new_config, "\tcloud_id = \"%s\"\n", personal_info.cloud_ID)
	fmt.Println("Please write the ID of the folder to operate under, if not specified by a given resource.")
	if _, err := fmt.Scan(&personal_info.folder_ID); err != nil {
		return personal_info, err
	}
	fmt.Fprintf(new_config, "\tfolder_id = \"%s\"\n", personal_info.folder_ID)

	zones := []string{"ru-central1-a", "ru-central1-b", "ru-central1-c", "ru-central1-d"}
	var err error
	for {
		fmt.Println("Please select the name of the default availability zone to operate under, if not specified by a given resource. Write only the number")
		personal_info.zone, err = choose_option(zones)
		if err != ErrInvalidInput {
			break
		}
	}
	fmt.Fprintf(new_config, "\tzone = \"%s\"\n", personal_info.zone)
	fmt.Fprintln(new_config, "}\n")
	return personal_info, nil
}

func main() {
	if _, err := os.Stat("yandex_cloud.tf"); err == nil {
		if err = os.Remove("yandex_cloud.tf"); err != nil {
			log.Fatalf("File deletion failed: %s", err.Error())
		}
	}

	config_name, err := select_config()
	if err != nil {
		log.Fatalf("The file could not be selected: %s", err.Error())
	}

	config, err := os.Open(config_name)
	if err != nil {
		log.Fatalf("File opening failed: %s", err.Error())
	}
	defer config.Close()

	general_info, network_adapters, vm_groups, err := scan_info(config)
	if err != nil {
		log.Fatalf("Error with information scanning: %s", err.Error())
	}

	new_config, err := os.Create("yandex_cloud.tf")
	if err != nil {
		log.Fatalf("File creation failed: %s", err.Error())
	}
	defer new_config.Close()

	personal_info, err := get_personal_info(new_config)
	if err != nil {
		log.Fatalf("Getting personal info failed: %s", err.Error())
	}

	yc_images, err := get_yc_images()
	if err != nil {
		log.Fatalf("Error with getting yc images: %s", err.Error())
	}

	network_name, err := network_creation(new_config)
	if err != nil {
		log.Fatalf("Error with network creation: %s", err.Error())
	}

	subnet_name, err := subnet_creation(new_config, network_name, personal_info.zone)
	if err != nil {
		log.Fatalf("Error with subnet creation: %s", err.Error())
	}

	conversion_of_resources(general_info, network_adapters, vm_groups, new_config, yc_images, network_name, subnet_name, personal_info)
}
