// Package main implements the basic logic of the converter.
package main

import (
	"bufio"
	"converter/resources"
	"converter/tools"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// select_config selects a configuration file from the available ones in the current directory
// based on the user's choice.
// The function returns the name of the selected file and any error encountered.
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
		return "", tools.ErrNoConfigurationFile
	} else {
		var config string
		for {
			fmt.Println("Please select the file to convert from this list. Write only the number.")
			config, err = tools.Choose_option(configs)
			if err != tools.ErrInvalidInput {
				break
			}
		}
		return config, nil
	}
}

// scan_info scans the information from the transmitted configuration file.
// The function returns any error encountered and 3 maps:
// general info about VMs in format map[name_of_VM](map[parameter_name]parameter_value),
// network adapters information for each VM in format map[name_of_VM](map[index_of_NA](map[parameter_name]parameter_value)),
// information about VM groups in format map[group_name][VM_names].
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

// get_personal_info asks the user the proper credentials and writes them to the transmitted configuration file.
// The function returns the structure with credentials and any error encountered.
func get_personal_info(new_config *os.File) (tools.Info, error) {
	var personal_info tools.Info
	fmt.Fprint(new_config, "terraform {\n\trequired_providers {\n\t\tyandex = {\n\t\t\tsource = \"yandex-cloud/yandex\"\n\t\t}\n\t}\n}\n\n")
	fmt.Fprintln(new_config, "provider \"yandex\" {")
	fmt.Println("Please write security token or IAM token used for authentication in Yandex.Cloud.")
	if _, err := fmt.Scan(&personal_info.OAuth_token); err != nil {
		return personal_info, err
	}
	fmt.Fprintf(new_config, "\ttoken = \"%s\"\n", personal_info.OAuth_token)
	fmt.Println("Please write the ID of the cloud to apply any resources to.")
	if _, err := fmt.Scan(&personal_info.Cloud_ID); err != nil {
		return personal_info, err
	}
	fmt.Fprintf(new_config, "\tcloud_id = \"%s\"\n", personal_info.Cloud_ID)
	fmt.Println("Please write the ID of the folder to operate under, if not specified by a given resource.")
	if _, err := fmt.Scan(&personal_info.Folder_ID); err != nil {
		return personal_info, err
	}
	fmt.Fprintf(new_config, "\tfolder_id = \"%s\"\n", personal_info.Folder_ID)

	zones := []string{"ru-central1-a", "ru-central1-b", "ru-central1-c", "ru-central1-d"}
	var err error
	for {
		fmt.Println("Please select the name of the default availability zone to operate under, if not specified by a given resource. Write only the number")
		personal_info.Zone, err = tools.Choose_option(zones)
		if err != tools.ErrInvalidInput {
			break
		}
	}
	fmt.Fprintf(new_config, "\tzone = \"%s\"\n", personal_info.Zone)
	fmt.Fprint(new_config, "}\n\n")
	return personal_info, nil
}

// conversion_of_resources is the main function of data conversion.
// The function accepts general info about VMs, network adapters information for each VM, information about VM groups, the configuration file to write to it,
// map with OS images for Yandex Cloud, the name of the network used, the name of the subnet used, the structure with credentials.
func conversion_of_resources(general_info map[string]map[string]string, network_adapters map[string](map[int](map[string]string)),
	vm_groups map[string][]string, new_config *os.File, yc_images map[string]string, network_name string, subnet_name string, personal_info tools.Info) {
	// Converting VM groups.
	if len(vm_groups) != 0 {
		if err := resources.Group_creation(general_info, network_adapters, vm_groups, new_config, yc_images, network_name, subnet_name, personal_info); err != nil {
			new_config.Close()
			os.Remove("yandex_cloud.tf")
			log.Fatalf("Error with resource group: %s", err.Error())
		}
	}

	created_folders := make([]string, 0, 1) // An array with the names of the created folders.
	var err error
	// Converting VMs.
	for reference_name, res := range general_info {
		if _, ok := res["group"]; !ok {
			folder_name, ok := res["basedir"]
			if ok {
				created_folders, err = resources.Folder_creation(created_folders, folder_name, new_config)
				if err != nil {
					new_config.Close()
					os.Remove("yandex_cloud.tf")
					log.Fatalf("Error with folder creation: %s", err.Error())
				}
			}
			if err := resources.VM_creation(reference_name, res, network_adapters, new_config, yc_images, subnet_name); err != nil {
				new_config.Close()
				os.Remove("yandex_cloud.tf")
				log.Fatalf("Error with resource virtual machine: %s", err.Error())
			}
		}
	}
}

// Main function for the converter.
func main() {
	// Deleting file "yandex_cloud.tf" if it exists.
	if _, err := os.Stat("yandex_cloud.tf"); err == nil {
		if err = os.Remove("yandex_cloud.tf"); err != nil {
			log.Fatalf("File deletion failed: %s", err.Error())
		}
	}

	// Selecting a configuration file for conversion.
	config_name, err := select_config()
	if err != nil {
		log.Fatalf("The file could not be selected: %s", err.Error())
	}

	// Opening the selected file.
	config, err := os.Open(config_name)
	if err != nil {
		log.Fatalf("File opening failed: %s", err.Error())
	}
	defer config.Close()

	// Scanning the information from the selected file.
	general_info, network_adapters, vm_groups, err := scan_info(config)
	if err != nil {
		log.Fatalf("Error with information scanning: %s", err.Error())
	}

	// Сreating a configuration file for Yandex Cloud with the name "yandex_cloud.tf".
	new_config, err := os.Create("yandex_cloud.tf")
	if err != nil {
		log.Fatalf("File creation failed: %s", err.Error())
	}
	defer new_config.Close()

	// Get the proper credentials.
	personal_info, err := get_personal_info(new_config)
	if err != nil {
		log.Fatalf("Getting personal info failed: %s", err.Error())
	}

	// Get map with OS images for Yandex Cloud.
	yc_images, err := tools.Get_yc_images()
	if err != nil {
		log.Fatalf("Error with getting yc images: %s", err.Error())
	}

	// Create network.
	network_name, err := resources.Network_creation(new_config)
	if err != nil {
		log.Fatalf("Error with network creation: %s", err.Error())
	}

	// Create subnet.
	subnet_name, err := resources.Subnet_creation(new_config, network_name, personal_info.Zone)
	if err != nil {
		log.Fatalf("Error with subnet creation: %s", err.Error())
	}

	// Сonverting resources and writing them to the configuration file for Yandex Cloud.
	conversion_of_resources(general_info, network_adapters, vm_groups, new_config, yc_images, network_name, subnet_name, personal_info)
}
