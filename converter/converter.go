package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
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

func scan_info(config *os.File) (map[string](map[string]string), error) {
	info := make(map[string](map[string]string))
	scanner := bufio.NewScanner(config)
	currVM := ""

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			splited := strings.Split(line, "=")
			info[currVM][strings.TrimSpace(splited[0])] = strings.TrimSpace(splited[1])
		} else if strings.Contains(line, "resource \"virtualbox_server\"") {
			splited := strings.Split(line, " ")
			currVM = splited[2]
			info[currVM] = make(map[string]string)
		}
	}

	err := scanner.Err()

	return info, err
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

// TODO: Write resource_network_interface
func resource_network_interface(new_config *os.File) {
	fmt.Fprint(new_config, "\tnetwork_interface {\n")
	fmt.Fprint(new_config, "\t\tsubnet_id = \"[192.168.0.0/16]\"\n")
	fmt.Fprint(new_config, "\t\tnat = true\n")
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

func conversion_of_resources(info map[string]map[string]string, new_config *os.File, yc_images map[string]string) {
	for reference_name, resources := range info {
		fmt.Fprintf(new_config, "resource \"yandex_compute_instance\" %s {\n", reference_name)

		optional_fields(resources, new_config)

		if err := resource_boot_disk(resources, new_config, yc_images); err != nil {
			new_config.Close()
			os.Remove("yandex_cloud.tf")
			log.Fatalf("Error with resource boot disk: %s", err.Error())
		}

		resource_network_interface(new_config)

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

	info, err := scan_info(config)
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

	conversion_of_resources(info, new_config, yc_images)
}
