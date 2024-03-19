package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrInvalidOperatingSystemName = errors.New("finding suitable yc_os_id failed")
var ErrNoConfigurationFile = errors.New("there is no configuration file for conversion in the directory")
var ErrOsIdNotDefined = errors.New("os_id must be specified")
var ErrCpusNotDefined = errors.New("cpus must be specified")
var ErrMemoryNotDefined = errors.New("memory must be specified")

func scan_info(config *os.File) (map[string](map[string]string), map[string](map[int](map[string]string)), error) {
	general_info := make(map[string](map[string]string))
	network_adapters := make(map[string](map[int](map[string]string)))
	scanner := bufio.NewScanner(config)
	currVM := ""
	network_adapter_count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			splited := strings.Split(line, "=")
			general_info[currVM][strings.TrimSpace(splited[0])] = strings.TrimSpace(splited[1])
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

	return general_info, network_adapters, err
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
		var result int
		for {
			fmt.Println("Please select the file to convert from this list. Write only the number.")
			for i, file := range configs {
				fmt.Printf("%d. %s\n", i+1, file)

			}
			if _, err = fmt.Scan(&result); err != nil {
				fmt.Println("Invalid input!")
				continue
			}
			if (1 <= result) && (result <= len(configs)) {
				break
			}
			fmt.Println("Invalid number!")
		}

		return configs[result-1], nil
	}
}

func copy_provider_info(new_config *os.File) error {
	source, err := os.Open("provider_info.txt")
	if err != nil {
		return err
	}

	_, err = io.Copy(new_config, source)

	source.Close()

	return err
}

func optional_fields(resources map[string]string, new_config *os.File) {
	count, ok := resources["count"]
	if ok {
		fmt.Fprintf(new_config, "\tcount = %s\n", count)
	}

	name, ok := resources["name"]
	if ok {
		fmt.Fprintf(new_config, "\tname = %s\n", name)
	}
}

func resource_boot_disk(resources map[string]string, new_config *os.File, yc_images map[string]string) error {
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

func resource_network_interface(network_adapters map[int]map[string]string, new_config *os.File, subnet_name string) {
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

func required_resources(resources map[string]string, new_config *os.File) error {
	fmt.Fprint(new_config, "\tresources {\n")

	cpus, ok := resources["cpus"]
	if !ok {
		return ErrCpusNotDefined
	}
	fmt.Fprintf(new_config, "\t\tcores = %s\n", cpus)

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
		mem++
	}

	fmt.Fprintf(new_config, "\t\tmemory = %d\n", mem)
	fmt.Fprint(new_config, "\t}\n")

	fmt.Fprint(new_config, "}\n\n")

	return nil
}

func network_creation(new_config *os.File) (string, error) {
	var network_name string
	fmt.Println("Please write a name for the network.")
	if _, err := fmt.Scan(&network_name); err != nil {
		return "", err
	}
	fmt.Fprintf(new_config, "resource \"yandex_vpc_network\" \"%s\" {\n", network_name)
	fmt.Fprintf(new_config, "\tname =  \"%s\"\n", network_name)
	fmt.Fprint(new_config, "}\n\n")
	return network_name, nil
}

func subnet_creation(new_config *os.File, network_name string) (string, error) {
	var subnet_name string
	fmt.Println("Please write a name for the subnet.")
	_, err := fmt.Scan(&subnet_name)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(new_config, "resource \"yandex_vpc_subnet\" \"%s\" {\n", subnet_name)
	fmt.Fprintf(new_config, "\tname =  \"%s\"\n", subnet_name)
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
	zones := []string{"ru-central1-a", "ru-central1-b", "ru-central1-c", "ru-central1-d"}
	var zone int
	for {
		fmt.Println("Please select the name of the Yandex.Cloud zone for this subnet. from this list. Write only the number")
		for i, zone := range zones {
			fmt.Printf("%d. %s\n", i+1, zone)

		}
		if _, err := fmt.Scan(&zone); err != nil {
			fmt.Println("Invalid input!")
			continue
		}
		if (1 <= zone) && (zone <= 4) {
			break
		}
		fmt.Println("Invalid number!")
	}
	fmt.Fprintf(new_config, "\tzone = \"%s\"\n", zones[zone-1])
	fmt.Fprintf(new_config, "\tnetwork_id = \"${yandex_vpc_network.%s.id}\"\n", network_name)
	fmt.Fprint(new_config, "}\n\n")
	return subnet_name, nil
}

func conversion_of_resources(general_info map[string]map[string]string, network_adapters map[string](map[int](map[string]string)), new_config *os.File, yc_images map[string]string, subnet_name string) {
	for reference_name, resources := range general_info {
		fmt.Fprintf(new_config, "resource \"yandex_compute_instance\" %s {\n", reference_name)

		optional_fields(resources, new_config)

		if err := resource_boot_disk(resources, new_config, yc_images); err != nil {
			new_config.Close()
			os.Remove("yandex_cloud.tf")
			log.Fatalf("Error with resource boot disk: %s", err.Error())
		}

		resource_network_interface(network_adapters[reference_name], new_config, subnet_name)

		if err := required_resources(resources, new_config); err != nil {
			new_config.Close()
			os.Remove("yandex_cloud.tf")
			log.Fatalf("Error with required resources: %s", err.Error())
		}
	}
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

	general_info, network_adapters, err := scan_info(config)
	if err != nil {
		log.Fatalf("Error with information scanning: %s", err.Error())
	}

	new_config, err := os.Create("yandex_cloud.tf")
	if err != nil {
		log.Fatalf("File creation failed: %s", err.Error())
	}
	defer new_config.Close()

	if err := copy_provider_info(new_config); err != nil {
		log.Fatalf("Copying data from \"provider_info.txt\" to \"yandex_cloud.tf\" failed: %s", err.Error())
	}

	yc_images, err := get_yc_images()
	if err != nil {
		log.Fatalf("Error with getting yc images: %s", err.Error())
	}

	network_name, err := network_creation(new_config)
	if err != nil {
		log.Fatalf("Error with network creation: %s", err.Error())
	}
	subnet_name, err := subnet_creation(new_config, network_name)
	if err != nil {
		log.Fatalf("Error with subnet creation: %s", err.Error())
	}
	conversion_of_resources(general_info, network_adapters, new_config, yc_images, subnet_name)
}
