package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	config, err := os.Open("virtualbox.tf")
	if err != nil {
		log.Fatal(err)
	}

	defer config.Close()

	scanner := bufio.NewScanner(config)

	info := make(map[string](map[string]string))

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

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat("yandex_cloud.tf"); err == nil {
		if err = os.Remove("yandex_cloud.tf"); err != nil {
			log.Fatal(err)
		}
	}

	new_config, err := os.Create("yandex_cloud.tf")
	if err != nil {
		log.Fatal(err)
	}
	defer new_config.Close()

	source, err := os.Open("provider_info.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	if _, err = io.Copy(new_config, source); err != nil {
		log.Fatal(err)
	}

	for reference_name, resources := range info {
		fmt.Fprintf(new_config, "resource \"yandex_compute_instance\" %s {\n", reference_name)

		name, ok := resources["name"]
		if ok {
			fmt.Fprintf(new_config, "\tname = %s\n", name)
		}

		fmt.Fprint(new_config, "\tboot_disk {\n")
		fmt.Fprint(new_config, "\t\tinitialize_params {\n")

		_, ok = resources["os_id"]
		if !ok {
			log.Fatal("os_id must be specified!")
		}
		//TODO: Write OS parsing
		fmt.Fprint(new_config, "\t\t\timage_id = \"fd8auu58m9ic4rtekngm\"\n")

		fmt.Fprint(new_config, "\t\t}\n")
		fmt.Fprint(new_config, "\t}\n")

		//TODO: vdi

		//TODO: Write network_interface parsing
		fmt.Fprint(new_config, "\tnetwork_interface {\n")
		fmt.Fprint(new_config, "\t\tsubnet_id = \"[192.168.0.0/16]\"\n")
		fmt.Fprint(new_config, "\t\tnat = true\n")
		fmt.Fprint(new_config, "\t}\n")

		fmt.Fprint(new_config, "\tresources {\n")

		cpus, ok := resources["cpus"]
		if !ok {
			log.Fatal("cpus must be specified!")
		}
		fmt.Fprintf(new_config, "\t\tcores = %s\n", cpus)

		memory, ok := resources["memory"]
		if !ok {
			log.Fatal("memory must be specified!")
		}
		mem, err := strconv.Atoi(memory)
		if err != nil {
			log.Fatal(err)
		}
		mem /= 1000
		if mem == 0 {
			mem++
		}
		fmt.Fprintf(new_config, "\t\tmemory = %d\n", mem)
		fmt.Fprint(new_config, "\t}\n")

		fmt.Fprint(new_config, "}\n\n")

	}
}
